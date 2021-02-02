package cni

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	enov1alpha1 "github.com/Nordix/eno/api/v1alpha1"

	"github.com/Nordix/eno/pkg/cni/cniconfig"

	"github.com/Nordix/eno/pkg/render"
)

// Ipam is an interface for IPAM Cnis
type Ipam interface {
	HandleIpam(ipamConf *cniconfig.IpamConfig, d *render.RenderData) error
}

// WhereAboutsIpam instance
type WhereAboutsIpam struct{}

type whereAboutsIpamJSON struct {
	Range  string  `json:"range,omitempty"`
	Type   string  `json:"type,omitempty"`
	Routes []route `json:"routes,omitempty"`
	Dns    *DNS    `json:"dns,omitempty"`
}

type route struct {
	Destination string `json:"dst,omitempty"`
	Gateway     string `json:"gw,omitempty"`
}

type DNS struct {
	Nameservers []string `json:"nameservers,omitempty"`
	Domain      string   `json:"domain,omitempty"`
	Search      []string `json:"search,omitempty"`
}

//NewWhereAboutsIpam creates instance of WhereAbouts cni
func NewWhereAboutsIpam() *WhereAboutsIpam {
	return &WhereAboutsIpam{}
}

// HandleIpam - Handles the whereabouts ipam cni case
func (ipam *WhereAboutsIpam) HandleIpam(ipamConf *cniconfig.IpamConfig, d *render.RenderData) error {
	var ipamObj whereAboutsIpamJSON
	ipamObj.Type = "whereabouts"
	err := ipam.populateRange(ipamConf, &ipamObj)
	if err != nil {
		return err
	}
	ipam.populateRoutes(ipamConf, &ipamObj)
	ipam.populateDNS(ipamConf, &ipamObj)
	marshalledConfig, err := json.Marshal(ipamObj)
	if err != nil {
		ipamConf.Log.Error(err, "Error marshalling ipam config")
		return err
	}
	d.Data["Ipam"] = string(marshalledConfig)
	ipamConf.Log.Info("Ipam config populated:", "config", d.Data["Ipam"])
	return nil
}

func (ipam *WhereAboutsIpam) populateRange(ipamConf *cniconfig.IpamConfig, ipamObj *whereAboutsIpamJSON) error {
	ipPool := ipamConf.Subnets[0].Spec.AllocationPool
	ipAddr := net.ParseIP(ipamConf.Subnets[0].Spec.Address)
	mask := ipamConf.Subnets[0].Spec.Mask
	if len(ipPool) > 1 {
		err := errors.New("more than one ip range not supported by WhereAbouts cni")
		ipamConf.Log.Error(err, "")
		return err
	}
	if len(ipPool) != 0 {
		ipamObj.Range = formatAllocationPool(ipPool[0]) + "/" + fmt.Sprint(mask)
	} else {
		ipamObj.Range = ipAddr.String() + "/" + fmt.Sprint(mask)
	}
	return nil
}

func (ipam *WhereAboutsIpam) populateRoutes(ipamConf *cniconfig.IpamConfig, ipamObj *whereAboutsIpamJSON) {
	if len(ipamConf.Routes[ipamConf.Subnets[0].GetName()]) > 0 {
		for _, cfgRoute := range ipamConf.Routes[ipamConf.Subnets[0].GetName()] {
			routePrefix := cfgRoute.Spec.Prefix + "/" + fmt.Sprint(cfgRoute.Spec.Mask)
			route := route{Destination: routePrefix, Gateway: cfgRoute.Spec.NextHop}
			ipamObj.Routes = append(ipamObj.Routes, route)
		}
	}
}

func (ipam *WhereAboutsIpam) populateDNS(ipamConf *cniconfig.IpamConfig, ipamObj *whereAboutsIpamJSON) {
	dns := ipamConf.Subnets[0].Spec.DNS
	if len(dns.Nameservers) == 0 && dns.Domain == "" && len(dns.Search) == 0 {
		return
	}
	dnsJSON := DNS{dns.Nameservers, dns.Domain, dns.Search}
	ipamObj.Dns = &dnsJSON
}

// Cnier is an interface for Cnis (e.g. ovs-cni, host-device-cni)
type Cnier interface {
	HandleCni(cniConf *cniconfig.CniConfig, d *render.RenderData) (string, error)
}

// OvsCni instance
type OvsCni struct{}

// NewOvsCni - creates an instance of OvsCni struct
func NewOvsCni() *OvsCni {
	return &OvsCni{}
}

// HandleCni - Handles the ovs-cni case
func (ovscni *OvsCni) HandleCni(cniConf *cniconfig.CniConfig, d *render.RenderData) (string, error) {
	manifestFolder := "ovs_netattachdef"
	//For VlanType=trunk we do not need to do anything
	switch cniConf.VlanType {
	case "access":
		if len(cniConf.VlanIds) != 1 {
			err := errors.New("Cannot use more than one Vlan for VlanType=access case")
			cniConf.Log.Error(err, "L2Services cannot contain more than one Vlan in VlanType=access case")
			return "", err
		}
		d.Data["AccessVlan"] = cniConf.VlanIds[0]
	case "selectivetrunk":
		tmpList := []string{}
		for _, vlanID := range cniConf.VlanIds {
			tmpStr := "{\"id\": " + strconv.Itoa(int(vlanID)) + "}"
			tmpList = append(tmpList, tmpStr)
		}
		d.Data["SelectiveVlan"] = "[" + strings.Join(tmpList, ",") + "]"
	case "trunk":
		cniConf.Log.Info("Transparent Trunk case in cluster level")
	}
	return manifestFolder, nil
}

// HostDevCni instance
type HostDevCni struct{}

// NewHostDevCni - creates an instance of HostDevCni struct
func NewHostDevCni() *HostDevCni {
	return &HostDevCni{}
}

// HandleCni - Handles the host-device-cni case
func (hdcni *HostDevCni) HandleCni(cniConf *cniconfig.CniConfig, d *render.RenderData) (string, error) {
	manifestFolder := "host-device_netattachdef"
	switch cniConf.VlanType {
	case "access":
		err := errors.New("Host-device cni does not support VlanType=access")
		cniConf.Log.Error(err, "Host-device VlanType error")
		return "", err
	case "selectivetrunk":
		err := errors.New("Host-device cni does not support VlanType=selectivetrunk")
		cniConf.Log.Error(err, "Host-device VlanType error")
		return "", err
	case "trunk":
		cniConf.Log.Info("Transparent Trunk case in Host-device cni")
	}
	return manifestFolder, nil
}

func formatAllocationPool(ipRange enov1alpha1.IPPool) string {
	return ipRange.Start + "-" + ipRange.End
}
