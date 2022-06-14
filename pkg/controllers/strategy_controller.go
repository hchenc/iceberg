package controller

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/hchenc/application/pkg/apis/app/v1beta1"
	servicemeshv1alpha2 "github.com/hchenc/iceberg/pkg/apis/servicemesh/v1alpha2"
	"github.com/hchenc/iceberg/pkg/constants"
	"github.com/hchenc/iceberg/pkg/controllers/filters"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func init() {
	RegisterReconciler("PatchApplication", SetUpStrategyReconcile)
}

type strategyReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func (s *strategyReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	strategy := &servicemeshv1alpha2.Strategy{}

	err := s.Get(ctx, req.NamespacedName, strategy)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		} else {
			log.Logger.WithFields(logrus.Fields{
				"strategy":  req.Name,
				"namespace": req.Namespace,
				"msg":       "failed to reconcile strategy",
			}).Error(err)
			return ctrl.Result{}, err
		}
	}

	if strategy.DeletionTimestamp != nil {
		return ctrl.Result{}, nil
	}

	log.Logger.WithFields(logrus.Fields{
		"action":    "PatchApplication",
		"strategy":  req.Name,
		"namespace": req.Namespace,
	}).Info("start to action")

	if appName, exist := strategy.Labels[constants.KubesphereAppName]; exist {
		application := &v1beta1.Application{}

		appNamespacedName := types.NamespacedName{
			Namespace: req.Namespace,
			Name:      appName,
		}
		err := s.Get(ctx, appNamespacedName, application)
		if err != nil {
			if errors.IsNotFound(err) {
				return ctrl.Result{}, nil
			} else {
				log.Logger.WithFields(logrus.Fields{
					"application": appName,
					"namespace":   req.Namespace,
					"msg":         "failed to get application",
				}).Error(err)
				return ctrl.Result{}, err
			}
		}

		application.Labels[constants.KubesphereAppVersion] = strategy.Spec.GovernorVersion
		if err := s.Update(ctx, application); err != nil {
			log.Logger.WithFields(logrus.Fields{
				"action": "PatchApplication",
			}).Error("failed to update application, ", err)
		}

		log.Logger.WithFields(logrus.Fields{
			"action": "PatchApplication",
		}).Info("finish to action")
		return reconcile.Result{}, nil
	} else {
		log.Logger.WithFields(logrus.Fields{
			"action": "PatchApplication",
		}).Error("failed to get application")
		return reconcile.Result{}, nil
	}
}

func (s *strategyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&servicemeshv1alpha2.Strategy{}).
		WithEventFilter(
			predicate.And(
				&filters.NamespaceUpdatePredicate{
					IncludeNamespaces: filters.DefaultIncludeNamespaces,
				},
				&strategyUpdatePredicate{},
			),
		).
		Complete(s)
}

func SetUpStrategyReconcile(mgr manager.Manager) {
	if err := (&strategyReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("PatchApplication"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		log.Fatalf("unable to create strategy controller for ", err)
	}
}

type strategyUpdatePredicate struct {
	filters.NamespaceUpdatePredicate
}

func (s strategyUpdatePredicate) Update(e event.UpdateEvent) bool {
	if references := e.ObjectOld.GetOwnerReferences(); references == nil {
		return false
	}
	return true
}
