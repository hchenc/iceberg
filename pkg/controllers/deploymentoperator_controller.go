package controller

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/hchenc/iceberg/pkg/constants"
	"github.com/hchenc/iceberg/pkg/controllers/filters"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"
)

func init() {
	RegisterReconciler("DeploymentToEnv", SetUpDeploymentReconcile)
}

type DeploymentOperatorReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func (d *DeploymentOperatorReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	deployment := &v1.Deployment{}

	err := d.Get(ctx, req.NamespacedName, deployment)
	if err != nil {
		if errors.IsNotFound(err) {
			d.Log.Info("it's a delete event")
		} else {
			log.Logger.WithFields(logrus.Fields{
				"deployment": req.Name,
				"namespace":  req.Namespace,
				"message":    "failed to reconcile deployment",
			}).Error(err)
		}
	} else {
		log.Logger.WithFields(logrus.Fields{
			"action": "DeploymentToEnv",
		}).Info("start to action")
		//inject reloader annotation
		if _, exist := deployment.Labels["reloader.efunds.com/auto"]; !exist {
			deployment.Labels["reloader.efunds.com/auto"] = "true"
			if err := d.Update(ctx, deployment); err != nil {
				log.Logger.WithFields(logrus.Fields{
					"action": "ReloaderInject",
				}).Error(err)
				return ctrl.Result{}, err
			}
		}

		//sync deployment to all environment(fat|uat|sit)
		_, err := deploymentGeneratorService.Add(deployment)
		if err != nil {
			log.Logger.WithFields(logrus.Fields{
				"event":    "create",
				"resource": "Deployment",
				"name":     deployment.Name,
				"result":   "failed",
				"error":    err.Error(),
			}).Errorf("deployment created failed, retry after %d second", RetryPeriod)
			return reconcile.Result{
				RequeueAfter: RetryPeriod * time.Second,
			}, err
		}
		log.Logger.WithFields(logrus.Fields{
			"event":    "create",
			"resource": "Deployment",
			"name":     deployment.Name,
			"result":   "success",
		}).Infof("finish to sync deployment %s", deployment.Name)
	}
	log.Logger.WithFields(logrus.Fields{
		"action": "DeploymentToEnv",
	}).Info("finish to action")
	return reconcile.Result{}, nil
}

func (d *DeploymentOperatorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.Deployment{}).
		WithEventFilter(
			predicate.And(
				&filters.NamespaceCreatePredicate{
					IncludeNamespaces: filters.DefaultIncludeNamespaces,
				},
				&filters.LabelCreatePredicate{
					Force: true,
					IncludeLabels: map[string]string{
						constants.KubesphereVersion: constants.KubesphereInitVersion,
					}},
			),
		).
		Complete(d)
}

func SetUpDeploymentReconcile(mgr manager.Manager) {
	if err := (&DeploymentOperatorReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("DeploymentToEnv"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		log.Fatalf("unable to create deployment controller for ", err)
	}
}
