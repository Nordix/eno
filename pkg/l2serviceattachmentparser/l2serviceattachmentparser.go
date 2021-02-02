package l2serviceattachmentparser

import (
	"errors"
	"fmt"

	"github.com/Nordix/eno/pkg/cni/cniconfig"
	"github.com/Nordix/eno/pkg/parser"

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
	srvAttResource  *enov1alpha1.L2ServiceAttachment
	cpResource      *enov1alpha1.ConnectionPoint
	l2srvResources  []*enov1alpha1.L2Service
	subnetResources []*enov1alpha1.Subnet
	routeResources  []*enov1alpha1.Route
	config          *config.Configuration
	cniMapping     map[string]cni.Cnier
	log             logr.Logger
}

// NewL2SrvAttParser - creates instance of L2SrvAttParser
func NewL2SrvAttParser(srvAttObj *enov1alpha1.L2ServiceAttachment, l2srvObjs []*enov1alpha1.L2Service,
	cpObj *enov1alpha1.ConnectionPoint, subnetObjs []*enov1alpha1.Subnet,
	routeObjs []*enov1alpha1.Route, c *config.Configuration, mc map[string]cni.Cnier, logger logr.Logger) *L2SrvAttParser {
	return &L2SrvAttParser{srvAttResource: srvAttObj,
		cpResource:      cpObj,
		l2srvResources:  l2srvObjs,
		subnetResources: subnetObjs,
		routeResources: routeObjs,
		config:          c,
		cniMapping:     mc,
		log:             logger}
}

// ParseL2ServiceAttachment - parses a L2ServiceAttachment Resource
func (sattp *L2SrvAttParser) ParseL2ServiceAttachment(d *render.RenderData) (string, error) {
	// Initiate Parsers
	cpParser := connectionpointparser.NewCpParser(sattp.cpResource, sattp.log)

	// Parse ConnectionPoint object
	cpParser.ParseConnectionPoint(d)

	cniToUse, err := sattp.pickCni()
	if err != nil {
		sattp.log.Error(err, "Error occurred while picking cni")
		return "", err
	}

	cniObj := sattp.cniMapping[cniToUse]
	cniConfigObj := sattp.instantiateCniConfig()
	manifestFolder, err := cniObj.HandleCni(cniConfigObj, d)
	if err != nil {
		sattp.log.Error(err, "Error occured while handling cni")
		return "", err
	}
	//TODO change picking one subnet if required
	if len(sattp.subnetResources) > 0 {
		sp := parser.NewSubnetParser(sattp.subnetResources[0], sattp.log)
		if err := sp.ValidateSubnet(); err != nil {
			return "", err
		}
		if err := sp.ValidateRoute(sattp.routeResources); err != nil {
			return "", err
		}

		cniObj := sp.PickIpamCni(sattp.routeResources)

		if err := cniObj.HandleIpam(d); err != nil {
			return "", err
		}
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
			if !common.SearchInSlice(sattp.srvAttResource.Spec.Implementation, cni.GetKernelSupportedCnis()) {
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

// instantiateCniConfig - Instantiate the CniConfig object to be used
func (sattp *L2SrvAttParser) instantiateCniConfig() *cniconfig.CniConfig {
	segIds := sattp.getSegIds()
	return cniconfig.NewCniConfig(segIds, sattp.srvAttResource.Spec.VlanType, sattp.log)
}

// getSegIds - returns a list with the segmentation Ids
func (sattp *L2SrvAttParser) getSegIds() []uint16 {
	var tmpslice []uint16
	for _, l2srvObj := range sattp.l2srvResources {
		tmpslice = append(tmpslice, l2srvObj.Spec.SegmentationID)
	}
	return tmpslice
}
