package l2serviceattachmentparser

import (
	enov1alpha1 "github.com/Nordix/eno/api/v1alpha1"
	"github.com/go-logr/logr"
)

// L2SrvParser instance
type L2SrvParser struct {
	l2srvResource *enov1alpha1.L2Service
	log           logr.Logger
}

// NewL2SrvParser - creates instance of L2SrvParser
func NewL2SrvParser(l2srvObj *enov1alpha1.L2Service, logger logr.Logger) *L2SrvParser {
	return &L2SrvParser{l2srvResource: l2srvObj,
		log: logger}
}
