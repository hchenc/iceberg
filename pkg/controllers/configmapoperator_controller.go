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

var (
	configmapAction = "ConfigMapToEnv"
)

func init() {
	RegisterReconciler(configmapAction, SetUpConfigMapReconcile)
}

type ConfigmapOperatorReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func (c *ConfigmapOperatorReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	configmap := &v1.ConfigMap{}

	err := c.Get(ctx, req.NamespacedName, configmap)
	if err != nil {
		if errors.IsNotFound(err) {
			c.Log.Info("receive delete event")
			return ctrl.Result{}, nil
		} else {
			log.Logger.WithFields(logrus.Fields{
				"configmap": req.Name,
				"namespace": req.Namespace,
				"message":   "failed to reconcile configmap",
			}).Error(err)
			return ctrl.Result{}, err
		}
	}

	if configmap.DeletionTimestamp != nil {
		return ctrl.Result{}, nil
	}

	log.Logger.WithFields(logrus.Fields{
		"action": configmapAction,
	}).Info("start to action")

	{
		//sync configmap to all environment(fat|uat|sit)
		_, err = configmapGeneratorService.Add(configmap)
		if err != nil {
			log.Logger.WithFields(logrus.Fields{
				"event":    "create",
				"resource": "ConfigMap",
				"name":     configmap.Name,
				"result":   "failed",
				"error":    err.Error(),
			}).Errorf("configmap sync to fat|uat|sit env failed, retry after %d second", RetryPeriod)
			return reconcile.Result{
				RequeueAfter: RetryPeriod * time.Second,
			}, err
		}
		log.Logger.WithFields(logrus.Fields{
			"event":    "create",
			"resource": "ConfigMap",
			"name":     configmap.Name,
			"result":   "success",
		}).Infof("finish to sync ConfigMap %c", configmap.Name)
	}
	log.Logger.WithFields(logrus.Fields{
		"action": configmapAction,
	}).Info("finish to action")
	return reconcile.Result{}, nil
}

func (c *ConfigmapOperatorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.ConfigMap{}).
		WithEventFilter(
			&filters.NamespaceCreatePredicate{
				IncludeNamespaces: filters.DefaultIncludeNamespaces,
			},
		).
		Complete(c)
}

func SetUpConfigMapReconcile(mgr manager.Manager) {
	if err := (&ConfigmapOperatorReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("ServiceToEnv"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		log.Fatalf("unable to create service controller for ", err)
	}
}
