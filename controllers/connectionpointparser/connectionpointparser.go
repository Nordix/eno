package connectionpointparser

import (
	enov1alpha1 "github.com/Nordix/eno/api/v1alpha1"
	"github.com/Nordix/eno/render"
)

// CpParser instance
type CpParser struct {
	cpResource *enov1alpha1.ConnectionPoint
}

// NewCpParser - creates instance of CpParser
func NewCpParser(cpObj *enov1alpha1.ConnectionPoint) *CpParser {
	return &CpParser{cpResource: cpObj}
}

// ParseConnectionPoint - parses a ConnectionPoint Resource
func (cpp *CpParser) ParseConnectionPoint(d *render.RenderData) {
	if cpp.cpResource.Spec.Type == "kernel" {
		d.Data["NetObjName"] = cpp.cpResource.Spec.InterfaceName
	} else {
		d.Data["NetObjName"] = cpp.cpResource.Spec.ResourceName
	}
}
