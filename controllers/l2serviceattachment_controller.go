/*


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

package controllers

import (
	"context"
	"fmt"
	"github.com/Nordix/eno/render"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"path/filepath"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	enov1alpha1 "github.com/Nordix/eno/api/v1alpha1"
	nettypes "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
	uns "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// L2ServiceAttachmentReconciler reconciles a L2ServiceAttachment object
type L2ServiceAttachmentReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=eno.k8s.io,resources=l2serviceattachments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=eno.k8s.io,resources=l2serviceattachments/status,verbs=get;update;patch

func (r *L2ServiceAttachmentReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("l2serviceattachment", req.NamespacedName)

	// Fetch the L2ServiceAttachment instance
	svc_att := &enov1alpha1.L2ServiceAttachment{}
	err := r.Get(ctx, req.NamespacedName, svc_att)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("L2ServiceAttachment resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get L2ServiceAttachment")
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}
	// Check if the NetAttachDef already exists, if not create a new one
	found := &nettypes.NetworkAttachmentDefinition{}
	err = r.Get(ctx, types.NamespacedName{Name: svc_att.Name, Namespace: svc_att.Namespace}, found)
	if err != nil && apierrors.IsNotFound(err) {
		// Define a new NetAttachDef
		log.Info("LIGOOO")
		log.Info(svc_att.Spec.L2Services[0])
		log.Info(svc_att.Spec.ConnectionPoint)
		net_att_def, err := r.defineNetAttachDef(ctx, log, svc_att)
		if err != nil {
			return ctrl.Result{}, err
		}
		//log.Info("Creating a new NetAttachDef", "NetAttachDef.Namespace", net_att_def.Namespace, "NetAttachDef.Name", net_att_def.Name)
		err = r.Create(ctx, net_att_def)
		if err != nil {
			//log.Error(err, "Failed to create new NetAttachDef", "NetAttachDef.Namespace", net_att_def.Namespace,
			//	"NetAttachDef.Name", net_att_def.Name)
			return ctrl.Result{}, err
		}
		// NetAttachDef created successfully - return
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get NetAttachDef")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

//func (r *L2ServiceAttachmentReconciler) defineNetAttachDef(ctx context.Context, log logr.Logger, s *enov1alpha1.L2ServiceAttachment) (*nettypes.NetworkAttachmentDefinition, error) {
func (r *L2ServiceAttachmentReconciler) defineNetAttachDef(ctx context.Context, log logr.Logger, s *enov1alpha1.L2ServiceAttachment) (*uns.Unstructured, error) {

	objs := []*uns.Unstructured{}
	data := render.MakeRenderData()

	cp_name := s.Spec.ConnectionPoint
	l2srv_list := s.Spec.L2Services
	vlan_type := s.Spec.VlanType
	log.Info("PAMEE_LIGOOO")
	log.Info(l2srv_list[0])
	log.Info(cp_name)
	log.Info(vlan_type)
	cp := &enov1alpha1.ConnectionPoint{}
	if err := r.Get(ctx, types.NamespacedName{Name: cp_name}, cp); err != nil {
		log.Error(err, "Failed to find ConnectionPoint", "ConnectionPoint.Name", cp_name)
		return nil, err
	}

	net_obj := ""
	if cp.Spec.Type == "kernel" {
		net_obj = cp.Spec.InterfaceName
	} else {
		net_obj = cp.Spec.ResourceName
	}

	seg_id_list := []uint16{}
	switch vlan_type {
	case "access":
		if len(l2srv_list) != 1 {
			err := errors.Errorf("Cannot use more than one L2Services for VlanType=access case")
			log.Error(err, "L2Services cannot contain more than one L2Services in VlanType=access case")
			return nil, err
		}

		l2srv_name := l2srv_list[0]
		l2srv := &enov1alpha1.L2Service{}
		if err := r.Get(ctx, types.NamespacedName{Name: l2srv_name}, l2srv); err != nil {
			log.Error(err, "Failed to find L2Service", "L2Service.Name", l2srv_name)
			return nil, err
		}
		seg_id_list = append(seg_id_list, l2srv.Spec.SegmentationID)
	}

	data.Data["NetAttachDefName"] = s.Name
	data.Data["NetAttachDefNamespace"] = s.Namespace
	data.Data["InterfaceName"] = net_obj
	data.Data["AccessVlan"] = seg_id_list[0]

	objs, err := render.RenderDir(filepath.Join("manifests", "ovs_netattachdef"), &data)

	if err != nil {
		return nil, err
	}
	log.Info("Dimitris")
	fmt.Printf("%+v\n", objs[0])
	ctrl.SetControllerReference(s, objs[0], r.Scheme)
	return objs[0], nil
	//return nil, err
}

func (r *L2ServiceAttachmentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&enov1alpha1.L2ServiceAttachment{}).
		Owns(&nettypes.NetworkAttachmentDefinition{}).
		Complete(r)
}
