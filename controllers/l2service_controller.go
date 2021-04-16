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
	"errors"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	enov1alpha1 "github.com/Nordix/eno/api/v1alpha1"
)

// L2ServiceReconciler reconciles a L2Service object
type L2ServiceReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=eno.k8s.io.k8s.io,resources=l2services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=eno.k8s.io.k8s.io,resources=l2services/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=eno.k8s.io,resources=l2bridgedomains,verbs=get;list;watch

func (r *L2ServiceReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	statusCPs := []string{}
	//_ = context.Background()
	//_ = r.Log.WithValues("l2service", req.NamespacedName)
	ctx := context.Background()
	log := r.Log.WithValues("l2service", req.NamespacedName)

	// Fetch the L2Service instance
	lTwoSvc := &enov1alpha1.L2Service{}
	err := r.Get(ctx, req.NamespacedName, lTwoSvc)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("L2Service resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get L2Service")
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}
	// Check if the L2BridgeDomain already exists, if not create a new one
	found := &enov1alpha1.L2BridgeDomain{}
	err = r.Get(ctx, types.NamespacedName{Name: lTwoSvc.Name}, found)
	if err != nil && apierrors.IsNotFound(err) {
		// Define a new L2BridgeDomain
		brDom, err := r.DefineBridgeDomain(ctx, log, lTwoSvc)
		if err != nil {
			if err := r.UpdateStatus(ctx, log, statusCPs, "error", err.Error(), lTwoSvc); err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, err
		}
		log.Info("Creating a new L2BridgeDomain", "L2BridgeDomain.Name", lTwoSvc.Name)
		err = r.Create(ctx, brDom)
		if err != nil {
			// Update status of L2Service resource with error phase
			if err := r.UpdateStatus(ctx, log, statusCPs, "error", err.Error(), lTwoSvc); err != nil {
				return ctrl.Result{}, err
			}
			log.Error(err, "Failed to create new L2BridgeDomain", "L2BridgeDomain.Name", lTwoSvc.Name)
			return ctrl.Result{}, err
		}
		// L2BridgeDomain created successfully - return
		// Update status of L2Service resource with pending phase
		if err := r.UpdateStatus(ctx, log, statusCPs, "pending", "Configuration pending", lTwoSvc); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		// Update status of L2Service resource with error phase
		if err := r.UpdateStatus(ctx, log, statusCPs, "error", err.Error(), lTwoSvc); err != nil {
			return ctrl.Result{}, err
		}
		log.Error(err, "Failed to get L2BridgeDomain")
		return ctrl.Result{}, err
	}

	candidateBrDom, err := r.DefineBridgeDomain(ctx, log, lTwoSvc)

	if err != nil {
		// Update status of L2ServiceAttachment resource with error phase
		if err := r.UpdateStatus(ctx, log, statusCPs, "error", err.Error(), lTwoSvc); err != nil {
			return ctrl.Result{}, err
		}
		log.Error(err, "Failed to define candidate L2BridgeDomain")
		return ctrl.Result{}, err
	}
	desiredCPs := candidateBrDom.Spec.ConnectionPoints
	actualCPs := found.Spec.ConnectionPoints

	if r.DiffDesiredActual(desiredCPs, actualCPs) {
		found.Spec.ConnectionPoints = desiredCPs
		err = r.Update(ctx, found)
		if err != nil {
			// Update status of L2Service resource with error phase
			if err := r.UpdateStatus(ctx, log, statusCPs, "error", err.Error(), lTwoSvc); err != nil {
				return ctrl.Result{}, err
			}
			log.Error(err, "Failed to update L2BridgeDomain", "L2BridgeDomain.Name", found.Name)
			return ctrl.Result{}, err
		}
		// Update status of L2Service resource with pending phase
		if err := r.UpdateStatus(ctx, log, statusCPs, "pending", "Configuration pending", lTwoSvc); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{Requeue: true}, nil
	}

	// Update status of L2Service resource
	if found.Status.Phase == "ready" {
		if err := r.UpdateStatus(ctx, log, desiredCPs, "ready", "Resources has been created", lTwoSvc); err != nil {
			return ctrl.Result{}, err
		}
	} else if found.Status.Phase == "error" {
		err := errors.New("L2BridgeDomain CR is in error state")
		if err := r.UpdateStatus(ctx, log, statusCPs, "error", err.Error(), lTwoSvc); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, err
	} else {
		// Update status of L2Service resource with pending phase
		if err := r.UpdateStatus(ctx, log, statusCPs, "pending", "Configuration pending", lTwoSvc); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	return ctrl.Result{}, nil
}

func (r *L2ServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&enov1alpha1.L2Service{}).
		Owns(&enov1alpha1.L2BridgeDomain{}).
		Watches(&source.Kind{Type: &enov1alpha1.L2ServiceAttachment{}}, &handler.EnqueueRequestsFromMapFunc{
			ToRequests: handler.ToRequestsFunc(func(o handler.MapObject) []reconcile.Request {
				var requests = []reconcile.Request{}
				svcAtt := o.Object.(*enov1alpha1.L2ServiceAttachment)
				for _, lTwoServiceName := range svcAtt.Spec.L2Services {
					requests = append(requests, reconcile.Request{
						NamespacedName: types.NamespacedName{
							Name: lTwoServiceName,
						},
					})
				}
				return requests
			}),
		}).
		WithEventFilter(lTwoServicePredicate()).
		Complete(r)
}
