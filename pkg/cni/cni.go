package cni

import (
	"errors"
	"strconv"
	"strings"

	enov1alpha1 "github.com/Nordix/eno/api/v1alpha1"
	"github.com/Nordix/eno/pkg/render"
	"github.com/go-logr/logr"
)

// Cnier is an interface for Cnis (e.g. ovs-cni, host-device-cni)
type Cnier interface {
	HandleCni(d *render.RenderData) error
}

// OvsCni instance
type OvsCni struct {
	L2srvResources []*enov1alpha1.L2Service
	VlanType       string
	Log            logr.Logger
}

// NewOvsCni - creates an instance of OvsCni struct
func NewOvsCni(l2srvObjs []*enov1alpha1.L2Service, typeOfVlan string, logger logr.Logger) *OvsCni {
	return &OvsCni{L2srvResources: l2srvObjs,
		VlanType: typeOfVlan,
		Log:      logger}
}

// HandleCni - Handles the ovs-cni case
func (ovscni *OvsCni) HandleCni(d *render.RenderData) error {

	//For VlanType=trunk we do not need to do anything
	switch ovscni.VlanType {
	case "access":
		if len(ovscni.L2srvResources) != 1 {
			err := errors.New("Cannot use more than one L2Services for VlanType=access case")
			ovscni.Log.Error(err, "L2Services cannot contain more than one L2Services in VlanType=access case")
			return err
		}
		d.Data["AccessVlan"] = ovscni.L2srvResources[0].Spec.SegmentationID
	case "selectivetrunk":
		tmpList := []string{}
		for _, l2srvObj := range ovscni.L2srvResources {
			tmpStr := "{\"id\": " + strconv.Itoa(int(l2srvObj.Spec.SegmentationID)) + "}"
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
