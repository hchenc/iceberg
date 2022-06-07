package controller

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/hchenc/iceberg/pkg/controllers/filters"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"
)

func init() {
	RegisterReconciler("ServiceToEnv", SetUpServiceReconcile)
}

type ServiceOperatorReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func (s *ServiceOperatorReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	service := &v1.Service{}

	err := s.Get(ctx, req.NamespacedName, service)
	if err != nil {
		if errors.IsNotFound(err) {
			s.Log.Info("it's a delete event")
		} else {
			log.Logger.WithFields(logrus.Fields{
				"service":   req.Name,
				"namespace": req.Namespace,
				"message":   "failed to reconcile service",
			}).Error(err)
		}
	} else {
		log.Logger.WithFields(logrus.Fields{
			"action": "ServiceToEnv",
		}).Info("start to action")
		//sync service to all environment(fat|uat|sit)
		_, err = serviceGeneratorService.Add(service)
		if err != nil {
			log.Logger.WithFields(logrus.Fields{
				"event":    "create",
				"resource": "Service",
				"name":     service.Name,
				"result":   "failed",
				"error":    err.Error(),
			}).Errorf("service sync to fat|uat|sit env failed, retry after %d second", RetryPeriod)
			return reconcile.Result{
				RequeueAfter: RetryPeriod * time.Second,
			}, err
		}
		log.Logger.WithFields(logrus.Fields{
			"event":    "create",
			"resource": "Service",
			"name":     service.Name,
			"result":   "success",
		}).Infof("finish to sync service %s", service.Name)
	}
	log.Logger.WithFields(logrus.Fields{
		"action": "ServiceToEnv",
	}).Info("finish to action")
	return reconcile.Result{}, nil
}

func (s *ServiceOperatorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.Service{}).
		WithEventFilter(&filters.NamespaceCreatePredicate{IncludeNamespaces: filters.DefaultIncludeNamespaces}).
		Complete(s)
}

func SetUpServiceReconcile(mgr manager.Manager) {
	if err := (&ServiceOperatorReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("ServiceToEnv"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		log.Fatalf("unable to create service controller for ", err)
	}
}
