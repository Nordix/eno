package cni

import (
	"errors"
	"strconv"
	"strings"

	"github.com/Nordix/eno/pkg/cni/cniconfig"

	"github.com/Nordix/eno/pkg/render"
)

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
