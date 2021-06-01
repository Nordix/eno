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
