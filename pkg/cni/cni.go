package cni

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-logr/logr"
	enov1alpha1 "github.com/Nordix/eno/api/v1alpha1"
	"net"
	"strconv"
	"strings"

	"github.com/Nordix/eno/pkg/cni/cniconfig"

	"github.com/Nordix/eno/pkg/render"
)

// Ipam is an interface for IPAM Cnis
type Ipam interface {
	HandleIpam(d *render.RenderData) error
}

// WhereAboutsIpam instance
type WhereAboutsIpam struct {
	Address   net.IP
	Mask      int16
	allocPool []enov1alpha1.IpPool
	Routes    []*enov1alpha1.Route
	Dns string
	log       logr.Logger
}

// WhereAboutsIpam instance
type WhereAboutsIpamJson struct {
	Range  string  `json:"range,omitempty"`
	Type   string  `json:"type,omitempty"`
	Routes []Route `json:"routes,omitempty"`
	Dns    string  `json:"dns,omitempty"`
}

type Route struct {
	Destination string `json:"dst,omitempty"`
	Gateway     string `json:"gw,omitempty"`
}

//NewWhereAboutsIpam creates instance of WhereAbouts cni
func NewWhereAboutsIpam(address net.IP, mask int16, ipPools []enov1alpha1.IpPool, routes []*enov1alpha1.Route, dns string, logger logr.Logger) *WhereAboutsIpam {
	return &WhereAboutsIpam{Address: address,
		Mask:      mask,
		allocPool: ipPools,
		Routes: routes,
		Dns: dns,
		log:       logger}
}

func (ipam *WhereAboutsIpam) HandleIpam(d *render.RenderData) error {
	if len(ipam.allocPool) > 1 {
		err := fmt.Errorf("more than one range not supported for WhereAbouts cni")
		ipam.log.Error(err, "")
		return err
	}

	AddrRange := formatAllocationPool(ipam.allocPool[0])
	var ipamObj WhereAboutsIpamJson
	ipamObj.Type = "whereabouts"
	if AddrRange != "" {
		ipamObj.Range = AddrRange + "/" + fmt.Sprint(ipam.Mask)
	} else {
		ipamObj.Range = ipam.Address.String() + "/" + fmt.Sprint(ipam.Mask)
	}

	//Populate Routes
	if len(ipam.Routes) > 0 {
		for _, cfgRoute := range ipam.Routes {
			routePrefix := cfgRoute.Spec.Prefix + "/" + fmt.Sprint(cfgRoute.Spec.Mask)
			route := Route{Destination: routePrefix, Gateway: cfgRoute.Spec.NextHop}
			ipamObj.Routes = append(ipamObj.Routes, route)
		}
	}

	//Populate dns json string
	ipamObj.Dns = ipam.Dns

	marshalledConfig, err := json.Marshal(ipamObj)
	if err != nil {
		ipam.log.Error(err, "Error marshalling ipam config")
		return err
	}
	d.Data["Ipam"] = string(marshalledConfig)
	ipam.log.Info("Ipam config populated:", "config", d.Data["Ipam"])
	return nil
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

func formatAllocationPool(ipRange enov1alpha1.IpPool) string {
	return ipRange.Start + "-" + ipRange.End
}