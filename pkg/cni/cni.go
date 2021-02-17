package cni

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	enov1alpha1 "github.com/Nordix/eno/api/v1alpha1"

	"github.com/Nordix/eno/pkg/cni/cniconfig"
)

// Ipam is an interface for IPAM Cnis
type Ipam interface {
	HandleIpam(ipamConf *cniconfig.IpamConfig, data map[string]interface{}) (string, error)
}

// WhereAboutsIpam instance
type WhereAboutsIpam struct{}

type route struct {
	Destination string
	Gateway     string
}

type DNS struct {
	Nameservers []string
	Domain      string
	Search      []string
}

//NewWhereAboutsIpam creates instance of WhereAbouts cni
func NewWhereAboutsIpam() *WhereAboutsIpam {
	return &WhereAboutsIpam{}
}

// HandleIpam - Handles the whereabouts ipam cni case
func (ipam *WhereAboutsIpam) HandleIpam(ipamConf *cniconfig.IpamConfig, data map[string]interface{}) (string, error) {
	manifestFile := "whereabouts.txt"
	err := ipam.populateRange(ipamConf, data)
	if err != nil {
		return "", err
	}
	ipam.populateRoutes(ipamConf, data)
	ipam.populateDNS(ipamConf, data)
	return manifestFile, nil
}

func (ipam *WhereAboutsIpam) populateRange(ipamConf *cniconfig.IpamConfig, data map[string]interface{}) error {
	ipPool := ipamConf.Subnets[0].Spec.AllocationPool
	ipAddr := net.ParseIP(ipamConf.Subnets[0].Spec.Address)
	mask := ipamConf.Subnets[0].Spec.Mask
	if len(ipPool) > 1 {
		err := errors.New("more than one ip range not supported by WhereAbouts cni")
		ipamConf.Log.Error(err, "")
		return err
	}
	//TODO: this if else also could be handled in template
	if len(ipPool) != 0 {
		data["Range"] = formatAllocationPool(ipPool[0]) + "/" + fmt.Sprint(mask)
	} else {
		data["Range"] = ipAddr.String() + "/" + fmt.Sprint(mask)
	}
	return nil
}

func (ipam *WhereAboutsIpam) populateRoutes(ipamConf *cniconfig.IpamConfig, data map[string]interface{}) {
	var routes []route
	if len(ipamConf.Routes[ipamConf.Subnets[0].GetName()]) > 0 {
		for _, cfgRoute := range ipamConf.Routes[ipamConf.Subnets[0].GetName()] {
			routePrefix := cfgRoute.Spec.Prefix + "/" + fmt.Sprint(cfgRoute.Spec.Mask)
			route := route{Destination: routePrefix, Gateway: cfgRoute.Spec.NextHop}
			routes = append(routes, route)
		}
	}
	if len(routes) > 0 {
		data["Routes"] = routes
	}
}

func (ipam *WhereAboutsIpam) populateDNS(ipamConf *cniconfig.IpamConfig, data map[string]interface{}) {
	dns := ipamConf.Subnets[0].Spec.DNS
	if len(dns.Nameservers) == 0 && dns.Domain == "" && len(dns.Search) == 0 {
		return
	}
	dnsIpam := DNS{dns.Nameservers, dns.Domain, dns.Search}
	data["Dns"] = dnsIpam
}

// Cnier is an interface for Cnis (e.g. ovs-cni, host-device-cni)
type Cnier interface {
	HandleCni(cniConf *cniconfig.CniConfig, data map[string]interface{}) (string, error)
}

// OvsCni instance
type OvsCni struct{}

// NewOvsCni - creates an instance of OvsCni struct
func NewOvsCni() *OvsCni {
	return &OvsCni{}
}

// HandleCni - Handles the ovs-cni case
func (ovscni *OvsCni) HandleCni(cniConf *cniconfig.CniConfig, data map[string]interface{}) (string, error) {
	manifestFile := "ovs.txt"
	//For VlanType=trunk we do not need to do anything
	switch cniConf.VlanType {
	case "access":
		if len(cniConf.VlanIds) != 1 {
			err := errors.New("Cannot use more than one Vlan for VlanType=access case")
			cniConf.Log.Error(err, "L2Services cannot contain more than one Vlan in VlanType=access case")
			return "", err
		}
		data["AccessVlan"] = cniConf.VlanIds[0]
	case "selectivetrunk":
		tmpList := []string{}
		for _, vlanID := range cniConf.VlanIds {
			tmpStr := "{\"id\": " + strconv.Itoa(int(vlanID)) + "}"
			tmpList = append(tmpList, tmpStr)
		}
		data["SelectiveVlan"] = "[" + strings.Join(tmpList, ",") + "]"
	case "trunk":
		cniConf.Log.Info("Transparent Trunk case in cluster level")
	}
	data["ResourcePrefix"] = "ovs-cni.network.kubevirt.io/"
	return manifestFile, nil
}

// HostDevCni instance
type HostDevCni struct{}

// NewHostDevCni - creates an instance of HostDevCni struct
func NewHostDevCni() *HostDevCni {
	return &HostDevCni{}
}

// HandleCni - Handles the host-device-cni case
func (hdcni *HostDevCni) HandleCni(cniConf *cniconfig.CniConfig, data map[string]interface{}) (string, error) {
	manifestFile := "host-device.txt"
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
	return manifestFile, nil
}

func formatAllocationPool(ipRange enov1alpha1.IPPool) string {
	return ipRange.Start + "-" + ipRange.End
}
