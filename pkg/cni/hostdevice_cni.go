package cni

import (
	l2SvcAttPars "github.com/Nordix/eno/pkg/l2serviceattachmentparser"
	"github.com/Nordix/eno/pkg/render"
	"github.com/Nordix/eno/pkg/framework"
	"github.com/go-logr/logr"
	"k8s.io/klog"
)

// HostDevCni instance
type HostDevCni struct {
	VlanType string
	Log      logr.Logger
}

const hostDeviceCniName = "host-device"

func init() {
	klog.Info("Registering Host Device CNI")
	if err := framework.Register(hostDeviceCniName, CreateHostDeviceCni); err != nil {
		klog.Fatalf("Cannot register %s: %v", hostDeviceCniName, err)
	}
}

// CreateHostDevCni - creates an instance of HostDevCni struct
func CreateHostDeviceCni() framework.Cnier {
	return &HostDevCni{}
}

// HandleCni - Handles the host-device-cni case
func (hdcni *HostDevCni) HandleCni(sattp *l2SvcAttPars.L2SrvAttParser, d *render.RenderData) error {
	//switch hdcni.VlanType {
	//case "access":
	//	err := errors.New("Host-device cni does not support VlanType=access")
	//	hdcni.Log.Error(err, "Host-device VlanType error")
	//	return err
	//case "selectivetrunk":
	//	err := errors.New("Host-device cni does not support VlanType=selectivetrunk")
	//	hdcni.Log.Error(err, "Host-device VlanType error")
	//	return err
	//case "trunk":
	//	hdcni.Log.Info("Transparent Trunk case in Host-device cni")
	//}
	return nil
}
