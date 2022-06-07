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
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"
)

var (
	secretAction = "SecretToEnv"
)

func init() {
	RegisterReconciler(secretAction, SetUpSecretReconcile)
}

type SecretOperatorReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func (s *SecretOperatorReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	secret := &v1.Secret{}

	err := s.Get(ctx, req.NamespacedName, secret)
	if err != nil {
		if errors.IsNotFound(err) {
			s.Log.Info("receive delete event")
			return ctrl.Result{}, nil
		} else {
			log.Logger.WithFields(logrus.Fields{
				"secret":    req.Name,
				"namespace": req.Namespace,
				"message":   "failed to reconcile secret",
			}).Error(err)
			return ctrl.Result{}, err
		}
	}

	if secret.DeletionTimestamp != nil {
		return ctrl.Result{}, nil
	}

	log.Logger.WithFields(logrus.Fields{
		"action": secretAction,
	}).Info("start to action")

	{
		//sync secret to all environment(fat|uat|sit)
		_, err = secretGeneratorService.Add(secret)
		if err != nil {
			log.Logger.WithFields(logrus.Fields{
				"event":    "create",
				"resource": "Secret",
				"name":     secret.Name,
				"result":   "failed",
				"error":    err.Error(),
			}).Errorf("secret sync to fat|uat|sit env failed, retry after %d second", RetryPeriod)
			return reconcile.Result{
				RequeueAfter: RetryPeriod * time.Second,
			}, err
		}
		log.Logger.WithFields(logrus.Fields{
			"event":    "create",
			"resource": "Secret",
			"name":     secret.Name,
			"result":   "success",
		}).Infof("finish to sync Secret %s", secret.Name)
	}
	log.Logger.WithFields(logrus.Fields{
		"action": secretAction,
	}).Info("finish to action")
	return reconcile.Result{}, nil
}

func (s *SecretOperatorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.Secret{}).
		WithEventFilter(
			predicate.And(
				&filters.NamespaceCreatePredicate{
					IncludeNamespaces: filters.DefaultIncludeNamespaces,
				},
				&filters.SecretCreatePredicate{},
			),
		).
		Complete(s)
}

func SetUpSecretReconcile(mgr manager.Manager) {
	if err := (&SecretOperatorReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("SecretToEnv"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		log.Fatalf("unable to create secret controller for ", err)
	}
}
