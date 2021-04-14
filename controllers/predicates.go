package controllers

import (
	enov1alpha1 "github.com/Nordix/eno/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

func ignoreStatusChangePredicate() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			// Ignore updates to CR status in which case metadata.Generation does not change
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
