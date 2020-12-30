package controllers

import (
	"context"

	enov1alpha1 "github.com/Nordix/eno/api/v1alpha1"
	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
)

// UpdateStatus - Update the Status of L2ServiceAttachment instance
func (r *L2ServiceAttachmentReconciler) UpdateStatus(ctx context.Context, req ctrl.Request, log logr.Logger, phase, message string) error {

	svcAttSt := &enov1alpha1.L2ServiceAttachment{}
	if err := r.Get(ctx, req.NamespacedName, svcAttSt); err != nil {
		log.Error(err, "Failed to get L2ServiceAttachment")
		return err
	}
	svcAttSt.Status.Phase = phase
	svcAttSt.Status.Message = message
	if err := r.Status().Update(ctx, svcAttSt); err != nil {
		log.Error(err, "Failed to update L2ServiceAttachment status")
		return err
	}
	return nil

}
