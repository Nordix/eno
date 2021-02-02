/*
Copyright 2021 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package framework

import (
	"errors"
	"fmt"
	enov1alpha1 "github.com/Nordix/eno/api/v1alpha1"
	"github.com/Nordix/eno/pkg/common"
	"github.com/Nordix/eno/pkg/config"
	"github.com/Nordix/eno/pkg/render"
	"github.com/Nordix/eno/pkg/l2serviceattachmentparser"
	"k8s.io/klog"
)

type cniManager struct {
	l2SvcAttParser *l2serviceattachmentparser.L2SrvAttParser
}

func CreateCniManager(l2SvcAttParser *l2serviceattachmentparser.L2SrvAttParser) *cniManager {
	return &cniManager{
		l2SvcAttParser: l2SvcAttParser,
	}
}

func (cm *cniManager) Execute(d *render.RenderData) {
	cniToUse, err := cm.pickCni()
	if err != nil {
		klog.Error("Error occurred while picking cni: ", err)
		return "", err
	}

	cniHandle, err := cm.getCniHandle(cniToUse)
	if err != nil {
		klog.Error("Error occurred while getting cni handle: ", err)
		return "", err
	}

	cniHandle.HandleCni(cm.l2SvcAttPars, d)
}

// pickCni - pick the CNI to be used for net-attach-def
func (cm *cniManager) pickCni() (string, error) {
	var cniToUse string
	// Default case - No CNI has been specified
	if cm.l2SvcAtt.Spec.Implementation == "" {
		if cm.l2SvcAtt.Spec.PodInterfaceType == "kernel" {
			cniToUse = cm.config.KernelCni
		} else {
			err := errors.New("We do not support default Cnis for PodInterfaceType=dpdk currently ")
			klog.Error("Error occurred while picking cni to use", err)
			return "", err
		}
	} else {
		if cm.l2SvcAtt.Spec.PodInterfaceType == "kernel" {
			if !common.SearchInSlice(cm.l2SvcAtt.Spec.Implementation, common.GetKernelSupportedCnis()) {
				err := fmt.Errorf(" %s cni is not supported currently", cm.l2SvcAtt.Spec.Implementation)
				klog.Error("Error occurred while picking cni to use: ", err)
				return "", err
			}
			cniToUse = cm.l2SvcAtt.Spec.Implementation
		} else {
			err := errors.New("We do not support Cnis for PodInterfaceType=dpdk currently ")
			klog.Error("Error occurred while picking cni to use: ", err)
			return "", err
		}
	}
	return cniToUse, nil
}

//instantiateCni - Instantiate the CNI object to be used
func (cm *cniManager) getCniHandle(cniName string) (Cnier, error) {
	return factory.createCniInstance(cniName)
}

// getSegIds - returns a list with the segmentation Ids
func (cm *cniManager) getSegIds() []uint16 {
	var tmpslice []uint16
	for _, l2srvObj := range cm.l2srvResources {
		tmpslice = append(tmpslice, l2srvObj.Spec.SegmentationID)
	}
	return tmpslice
}