package cniconfig

import (
	enov1alpha1 "github.com/Nordix/eno/api/v1alpha1"
	"github.com/go-logr/logr"
)

type CniConfig struct {
	VlanIds  []uint16
	VlanType string
	Log      logr.Logger
}

// NewCniConfig - creates an instance of CniConfig struct
func NewCniConfig(vlans []uint16, typeOfVlan string, logger logr.Logger) *CniConfig {
	return &CniConfig{VlanIds: vlans,
		VlanType: typeOfVlan,
		Log:      logger}
}

type IpamConfig struct {
	Subnets []*enov1alpha1.Subnet
	Routes  map[string][]*enov1alpha1.Route
	Log     logr.Logger
}

// NewIpamCniConfig - creates an instance of IpamCniConfig struct
func NewIpamConfig(subnets []*enov1alpha1.Subnet, routes map[string][]*enov1alpha1.Route, logger logr.Logger) *IpamConfig {
	return &IpamConfig{Subnets: subnets,
		Routes: routes,
		Log:    logger}
}
