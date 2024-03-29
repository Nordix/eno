package cniconfig

import (
	enov1alpha1 "github.com/Nordix/eno/api/v1alpha1"
	"github.com/go-logr/logr"
)

// CniConfig contains data required for cni's
type CniConfig struct {
	VlanIds          []uint16
	VlanType         string
	PodInterfaceType string
	CniOpts          map[string]interface{}
	Log              logr.Logger
}

// NewCniConfig - creates an instance of CniConfig struct
func NewCniConfig(vlans []uint16, typeOfVlan string, typeOfPodInterface string, cniOptions map[string]interface{}, logger logr.Logger) *CniConfig {
	return &CniConfig{VlanIds: vlans,
		VlanType:         typeOfVlan,
		PodInterfaceType: typeOfPodInterface,
		CniOpts:          cniOptions,
		Log:              logger}
}

// IpamConfig contains data required by Ipam cni
type IpamConfig struct {
	Subnets []*enov1alpha1.Subnet
	Routes  map[string][]*enov1alpha1.Route
	Log     logr.Logger
}

// NewIpamConfig - creates an instance of IpamCniConfig struct
func NewIpamConfig(subnets []*enov1alpha1.Subnet, routes map[string][]*enov1alpha1.Route, logger logr.Logger) *IpamConfig {
	return &IpamConfig{Subnets: subnets,
		Routes: routes,
		Log:    logger}
}
