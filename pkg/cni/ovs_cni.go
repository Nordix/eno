package cni

import (
	"errors"
	"github.com/Nordix/eno/pkg/framework"
	"k8s.io/klog"
	"strconv"
	"strings"

	"github.com/Nordix/eno/pkg/render"
	"github.com/go-logr/logr"
)

// OvsCni instance
type OvsCni struct {
	VlanIds  []uint16
	VlanType string
	Log      logr.Logger
}

const ovsCniName = "OvsCNI"

func init() {
	klog.Info("Registering Ovs CNI")
	if err := framework.Register(ovsCniName, CreateOvsCni); err != nil {
		klog.Fatalf("Cannot register %s: %v", ovsCniName, err)
	}
}


// CreateOvsCni - creates an instance of OvsCni struct
func CreateOvsCni() framework.Cnier {
	return &OvsCni{}
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


