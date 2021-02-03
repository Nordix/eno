package cniconfig

import (
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
