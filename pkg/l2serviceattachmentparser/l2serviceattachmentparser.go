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
	"github.com/Nordix/eno/pkg/framework"
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

	cniManager := framework.CreateCniManager(sattp)

	cniManager.Execute()

	return manifestFolder, nil
}

// getSegIds - returns a list with the segmentation Ids
func (sattp *L2SrvAttParser) getSegIds() []uint16 {
	var tmpslice []uint16
	for _, l2srvObj := range sattp.l2srvResources {
		tmpslice = append(tmpslice, l2srvObj.Spec.SegmentationID)
	}
	return tmpslice
}
