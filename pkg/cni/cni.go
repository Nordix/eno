package cni

import (
	"errors"
	"strconv"
	"strings"

	"github.com/Nordix/eno/pkg/render"
	"github.com/go-logr/logr"
)

// Cnier is an interface for Cnis (e.g. ovs-cni, host-device-cni)
type Cnier interface {
	HandleCni(d *render.RenderData) error
}

// OvsCni instance
type OvsCni struct {
	VlanIds  []uint16
	VlanType string
	Log      logr.Logger
}

// NewOvsCni - creates an instance of OvsCni struct
func NewOvsCni(vlans []uint16, typeOfVlan string, logger logr.Logger) *OvsCni {
	return &OvsCni{VlanIds: vlans,
		VlanType: typeOfVlan,
		Log:      logger}
}

// HandleCni - Handles the ovs-cni case
func (ovscni *OvsCni) HandleCni(d *render.RenderData) error {

	//For VlanType=trunk we do not need to do anything
	switch ovscni.VlanType {
	case "access":
		if len(ovscni.VlanIds) != 1 {
			err := errors.New("Cannot use more than one Vlan for VlanType=access case")
			ovscni.Log.Error(err, "L2Services cannot contain more than one Vlan in VlanType=access case")
			return err
		}
		d.Data["AccessVlan"] = ovscni.VlanIds[0]
	case "selectivetrunk":
		tmpList := []string{}
		for _, vlanId := range ovscni.VlanIds {
			tmpStr := "{\"id\": " + strconv.Itoa(int(vlanId)) + "}"
			tmpList = append(tmpList, tmpStr)
		}
		d.Data["SelectiveVlan"] = "[" + strings.Join(tmpList, ",") + "]"
	case "trunk":
		ovscni.Log.Info("Transparent Trunk case in cluster level")
	}
	return nil
}

// HostDevCni instance
type HostDevCni struct {
	VlanType string
	Log      logr.Logger
}

// NewHostDevCni - creates an instance of HostDevCni struct
func NewHostDevCni(typeOfVlan string, logger logr.Logger) *HostDevCni {
	return &HostDevCni{VlanType: typeOfVlan,
		Log: logger}
}

// HandleCni - Handles the host-device-cni case
func (hdcni *HostDevCni) HandleCni(d *render.RenderData) error {

	switch hdcni.VlanType {
	case "access":
		err := errors.New("Host-device cni does not support VlanType=access")
		hdcni.Log.Error(err, "Host-device VlanType error")
		return err
	case "selectivetrunk":
		err := errors.New("Host-device cni does not support VlanType=selectivetrunk")
		hdcni.Log.Error(err, "Host-device VlanType error")
		return err
	case "trunk":
		hdcni.Log.Info("Transparent Trunk case in Host-device cni")
	}
	return nil
}
