package controllers

import (
	"context"
	"path/filepath"

	enov1alpha1 "github.com/Nordix/eno/api/v1alpha1"
	"github.com/Nordix/eno/pkg/l2serviceattachmentparser"
	"github.com/Nordix/eno/pkg/render"
	"github.com/go-logr/logr"
	uns "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
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

// DefineNetAttachDef - Create and returns an Unstructured net-attach-def
func (r *L2ServiceAttachmentReconciler) DefineNetAttachDef(ctx context.Context, log logr.Logger, s *enov1alpha1.L2ServiceAttachment) (*uns.Unstructured, error) {

	objs := []*uns.Unstructured{}
	l2srvObjs := []*enov1alpha1.L2Service{}
	data := render.MakeRenderData()

	cpName := s.Spec.ConnectionPoint
	l2srvNames := s.Spec.L2Services

	// Get the ConnectionPoint resource
	cp := &enov1alpha1.ConnectionPoint{}
	if err := r.Get(ctx, types.NamespacedName{Name: cpName}, cp); err != nil {
		log.Error(err, "Failed to find ConnectionPoint", "ConnectionPoint.Name", cpName)
		return nil, err
	}
	// Get one or more L2Service resources
	for _, l2srvName := range l2srvNames {
		tempObj := &enov1alpha1.L2Service{}
		if err := r.Get(ctx, types.NamespacedName{Name: l2srvName}, tempObj); err != nil {
			log.Error(err, "Failed to find L2Service", "L2Service.Name", l2srvName)
			return nil, err
		}
		l2srvObjs = append(l2srvObjs, tempObj)
	}
	// Initiate L2ServiceAttachment Parser
	l2srvAttParser := l2serviceattachmentparser.NewL2SrvAttParser(s, l2srvObjs, cp, r.Config, log)
	// Parse the resources and fill the data
	manifestFolder, err := l2srvAttParser.ParseL2ServiceAttachment(&data)
	if err != nil {
		log.Error(err, "Error occurred during parsing the L2ServiceAttachment")
		return nil, err
	}

	data.Data["NetAttachDefName"] = s.Name
	data.Data["NetAttachDefNamespace"] = s.Namespace

	objs, err = render.RenderDir(filepath.Join("manifests", manifestFolder), &data)
	if err != nil {
		return nil, err
	}

	ctrl.SetControllerReference(s, objs[0], r.Scheme)
	return objs[0], nil
}
