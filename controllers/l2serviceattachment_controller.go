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
	"k8s.io/apimachinery/pkg/types"
        "k8s.io/apimachinery/pkg/api/errors"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	enov1alpha1 "github.com/Nordix/eno/api/v1alpha1"
        nettypes "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
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
		if errors.IsNotFound(err) {
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
	if err != nil && errors.IsNotFound(err) {
		// Define a new NetAttachDef
		net_att_def, err := r.defineNetAttachDef(svc_att)
		if err != nil {
		    return ctrl.Result{}, err
		}
		//log.Info("Creating a new NetAttachDef", "NetAttachDef.Namespace", net_att_def.Namespace, "NetAttachDef.Name", net_att_def.Name)
		//err = r.Create(ctx, net_att_def)
		//if err != nil {
		//	log.Error(err, "Failed to create new NetAttachDef", "NetAttachDef.Namespace", net_att_def.Namespace,
                //            "NetAttachDef.Name", net_att_def.Name,)
		//	return ctrl.Result{}, err
		//}
		// NetAttachDef created successfully - return
		return ctrl.Result{}, nil
	} else if err != nil {
		log.Error(err, "Failed to get NetAttachDef")
		return ctrl.Result{}, err
	}



	return ctrl.Result{}, nil
}

func (r *L2ServiceAttachmentReconciler) defineNetAttachDef(s *enov1alpha1.L2ServiceAttachment) (*nettypes.NetworkAttachmentDefinition, error) {
	cp_name := s.Spec.ConnectionPoint
	l2srv_list := s.Spec.L2Services
	vlan_type := s.Spec.VlanType

        cp := &enov1alpha1.ConnectionPoint{}
	if err := r.Get(ctx, types.NamespacedName{Name: cp_name}, cp); err != nil{
	    log.Error(err, "Failed to find ConnectionPoint", "ConnectionPoint.Name", cp_name)
	    return nil, err
	}

	net_obj := ""
        if *cp.Spec.Type == "kernel" {
            net_obj = *cp.Spec.InterfaceName
        }
        else {
            net_obj = *cp.Spec.ResourceName
        }

        seg_id_list := []uint16{}
	switch vlan_type {
	case "access":
	    if len(l2srv_list) != 1 {
                log.Error("L2Services cannot contain more than one L2Services in VlanType=access case")
		err := errors.Error("Cannot use more than one L2Services for VlanType=access case")
		return nil, err
	    }

	    l2srv_name := l2srv_list[0]
	    l2srv := &enov1alpha1.L2Service{}
	    if err := r.Get(ctx, types.NamespacedName{Name: l2srv_name}, l2srv); err != nil{
                log.Error(err, "Failed to find L2Service", "L2Service.Name", l2srv_name)
                return nil, err
            }
	    seg_id_list = append(seg_id_list, *l2srv.Spec.SegmentationID)
	}

        log.Info("AUTO ",cp)
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image:   "memcached:1.4.36-alpine",
						Name:    "memcached",
						Command: []string{"memcached", "-m=64", "-o", "modern", "-v"},
						Ports: []corev1.ContainerPort{{
							ContainerPort: 11211,
							Name:          "memcached",
						}},
					}},
				},
			},
		},
	}
	// Set Memcached instance as the owner and controller
	ctrl.SetControllerReference(m, dep, r.Scheme)
	return dep
}

func (r *L2ServiceAttachmentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&enov1alpha1.L2ServiceAttachment{}).
                Owns(&nettypes.NetworkAttachmentDefinition{}).
		Complete(r)
}
