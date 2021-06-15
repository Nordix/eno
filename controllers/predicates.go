package controllers

import (
	enov1alpha1 "github.com/Nordix/eno/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

func lTwoServiceAttPredicate() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			// Ignore updates to CR status in which case metadata.Generation does not change
			// Except for update in L2Service objects

			lTwoSvc, ok := e.ObjectNew.(*enov1alpha1.L2Service)
			if ok {
				if e.MetaOld.GetGeneration() != e.MetaNew.GetGeneration() {
					return true
				} else {
					if lTwoSvc.Status.Phase == "error" {
						return true
					}
				}
			}

			return e.MetaOld.GetGeneration() != e.MetaNew.GetGeneration()
		},
	}
}

func lTwoServicePredicate() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			// Ignore updates to CR status in which case metadata.Generation does not change
			// Except for update in L2BridgeDomain objects

			_, ok := e.ObjectOld.(*enov1alpha1.L2BridgeDomain)
			if ok {
				return true
			}

			return e.MetaOld.GetGeneration() != e.MetaNew.GetGeneration()
		},
	}
}
