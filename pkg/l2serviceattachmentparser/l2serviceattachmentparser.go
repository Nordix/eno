package l2serviceattachmentparser

import (
	"errors"
	"fmt"

	enov1alpha1 "github.com/Nordix/eno/api/v1alpha1"
	"github.com/Nordix/eno/pkg/cni"
	"github.com/Nordix/eno/pkg/common"
	"github.com/Nordix/eno/pkg/config"
	"github.com/Nordix/eno/pkg/connectionpointparser"
	"github.com/Nordix/eno/pkg/render"
	"github.com/go-logr/logr"
)

// L2SrvAttParser instance
type L2SrvAttParser struct {
	srvAttResource *enov1alpha1.L2ServiceAttachment
	cpResource     *enov1alpha1.ConnectionPoint
	l2srvResources []*enov1alpha1.L2Service
	config         *config.Configuration
	log            logr.Logger
}

// NewL2SrvAttParser - creates instance of L2SrvAttParser
func NewL2SrvAttParser(srvAttObj *enov1alpha1.L2ServiceAttachment, l2srvObjs []*enov1alpha1.L2Service,
	cpObj *enov1alpha1.ConnectionPoint, c *config.Configuration, logger logr.Logger) *L2SrvAttParser {
	return &L2SrvAttParser{srvAttResource: srvAttObj,
		cpResource:     cpObj,
		l2srvResources: l2srvObjs,
		config:         c,
		log:            logger}
}

// ParseL2ServiceAttachment - parses a L2ServiceAttachment Resource
func (sattp *L2SrvAttParser) ParseL2ServiceAttachment(d *render.RenderData) (string, error) {
	var manifestFolder string
	// Initiate Parsers
	cpParser := connectionpointparser.NewCpParser(sattp.cpResource, sattp.log)

	// Parse ConnectionPoint object
	cpParser.ParseConnectionPoint(d)

	cniToUse, err := sattp.pickCni()
	if err != nil {
		sattp.log.Error(err, "Error occured while picking cni")
		return "", err
	}

	manifestFolder, cniObj := sattp.instantiateCni(cniToUse)

	if err := cniObj.HandleCni(d); err != nil {
		sattp.log.Error(err, "Error occured while handling cni")
		return "", err
	}

	return manifestFolder, nil
}

// pickCni - pick the CNI to be used for net-attach-def
func (sattp *L2SrvAttParser) pickCni() (string, error) {
	var cniToUse string

	// Default case - No CNI has been specified
	if sattp.srvAttResource.Spec.Implementation == "" {
		if sattp.srvAttResource.Spec.PodInterfaceType == "kernel" {
			cniToUse = sattp.config.KernelCni
		} else {
			err := errors.New("We do not support default Cnis for PodInterfaceType=dpdk currenlty ")
			sattp.log.Error(err, "Error occured while picking cni to use")
			return "", err
		}
	} else {
		if sattp.srvAttResource.Spec.PodInterfaceType == "kernel" {
			if !common.SearchInSlice(sattp.srvAttResource.Spec.Implementation, common.GetKernelSupportedCnis()) {
				err := fmt.Errorf(" %s cni is not supported currently", sattp.srvAttResource.Spec.Implementation)
				sattp.log.Error(err, "Error occured while picking cni to use")
				return "", err
			}
			cniToUse = sattp.srvAttResource.Spec.Implementation
		} else {
			err := errors.New("We do not support Cnis for PodInterfaceType=dpdk currenlty ")
			sattp.log.Error(err, "Error occured while picking cni to use")
			return "", err
		}
	}
	return cniToUse, nil
}

// instantiateCni - Instantiate the CNI object to be used
func (sattp *L2SrvAttParser) instantiateCni(cniToUse string) (string, cni.Cnier) {
	var cniObj cni.Cnier
	var manifestFolder string
	switch cniToUse {
	case "ovs":
		cniObj = cni.NewOvsCni(sattp.l2srvResources, sattp.srvAttResource.Spec.VlanType, sattp.log)
		manifestFolder = "ovs_netattachdef"
	case "host-device":
		cniObj = cni.NewHostDevCni(sattp.srvAttResource.Spec.VlanType, sattp.log)
		manifestFolder = "host-device_netattachdef"
	}
	return manifestFolder, cniObj

}
