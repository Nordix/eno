package connectionpointparser

import (
	enov1alpha1 "github.com/Nordix/eno/api/v1alpha1"
	"github.com/go-logr/logr"
)

// CpParser instance
type CpParser struct {
	cpResource *enov1alpha1.ConnectionPoint
	log        logr.Logger
}

// NewCpParser - creates instance of CpParser
func NewCpParser(cpObj *enov1alpha1.ConnectionPoint, logger logr.Logger) *CpParser {
	return &CpParser{cpResource: cpObj,
		log: logger}
}

// ParseConnectionPoint - parses a ConnectionPoint Resource
func (cpp *CpParser) ParseConnectionPoint(data map[string]interface{}) {
	if cpp.cpResource.Spec.Type == "kernel" {
		data["NetObjName"] = cpp.cpResource.Spec.InterfaceName
	} else {
		data["NetObjName"] = cpp.cpResource.Spec.ResourceName
	}
}
