package l2serviceattachmentparser

import (
	"errors"
	"strconv"
	"strings"

	enov1alpha1 "github.com/Nordix/eno/api/v1alpha1"
	"github.com/Nordix/eno/pkg/connectionpointparser"
	"github.com/Nordix/eno/pkg/render"
	"github.com/go-logr/logr"
)

// L2SrvAttParser instance
type L2SrvAttParser struct {
	srvAttResource *enov1alpha1.L2ServiceAttachment
	cpResource     *enov1alpha1.ConnectionPoint
	l2srvResources []*enov1alpha1.L2Service
	log            logr.Logger
}

// NewL2SrvAttParser - creates instance of L2SrvAttParser
func NewL2SrvAttParser(srvAttObj *enov1alpha1.L2ServiceAttachment, l2srvObjs []*enov1alpha1.L2Service, cpObj *enov1alpha1.ConnectionPoint, logger logr.Logger) *L2SrvAttParser {
	return &L2SrvAttParser{srvAttResource: srvAttObj,
		cpResource:     cpObj,
		l2srvResources: l2srvObjs,
		log:            logger}
}

// ParseL2ServiceAttachment - parses a L2ServiceAttachment Resource
func (sattp *L2SrvAttParser) ParseL2ServiceAttachment(d *render.RenderData) (string, error) {
	var manifestFolder string
	// Initiate Parsers
	cpParser := connectionpointparser.NewCpParser(sattp.cpResource, sattp.log)

	// Parse ConnectionPoint object
	cpParser.ParseConnectionPoint(d)

	cniToUse, err := sattp.pickCni(sattp.srvAttResource.Spec.Implementation)
	if err != nil {
		sattp.log.Error(err, "Error occured while picking cni")
		return "", err
	}

	// Parse the L2Services objects
	switch cniToUse {
	case "ovs":
		if err := sattp.handleOvsCniCase(d); err != nil {
			sattp.log.Error(err, "Error occured while handling ovs-cni case")
			return "", err
		}
		manifestFolder = "ovs_netattachdef"
	case "host-device":
		manifestFolder = "host-device_netattachdef"
	}

	return manifestFolder, nil
}

// pickCni - pick the CNI to be used for net-attach-def
func (sattp *L2SrvAttParser) pickCni(cni string) (string, error) {
	// TODO: Add support for default CNIs
	if cni == "" {
		err := errors.New("Please define a CNI to use. We do not support default CNIs yet")
		return "", err
	}
	return sattp.srvAttResource.Spec.Implementation, nil
}

// handleOvsCniCase - Handles the ovs-cni case
func (sattp *L2SrvAttParser) handleOvsCniCase(d *render.RenderData) error {

	//For VlanType=trunk we do not need to do anything
	switch sattp.srvAttResource.Spec.VlanType {
	case "access":
		if len(sattp.l2srvResources) != 1 {
			err := errors.New("Cannot use more than one L2Services for VlanType=access case")
			sattp.log.Error(err, "L2Services cannot contain more than one L2Services in VlanType=access case")
			return err
		}
		d.Data["AccessVlan"] = sattp.l2srvResources[0].Spec.SegmentationID
	case "selectivetrunk":
		tmpList := []string{}
		for _, l2srvObj := range sattp.l2srvResources {
			tmpStr := "{\"id\": " + strconv.Itoa(int(l2srvObj.Spec.SegmentationID)) + "}"
			tmpList = append(tmpList, tmpStr)
		}
		d.Data["SelectiveVlan"] = "[" + strings.Join(tmpList, ",") + "]"
	case "trunk":
		sattp.log.Info("Transparent Trunk case in cluster level")
	}
	return nil
}
