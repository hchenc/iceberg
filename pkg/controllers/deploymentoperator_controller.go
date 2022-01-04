package controller

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/hchenc/iceberg/pkg/constants"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"strings"
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

type deploymentPredicate struct {
}

func (d deploymentPredicate) Create(e event.CreateEvent) bool {
	name := e.Object.GetNamespace()
	labels := e.Object.GetLabels()
	if strings.Contains(name, "system") || strings.Contains(name, "kube") {
		return false
	}

	if strings.Contains(name, "sit") || strings.Contains(name, "fat") || strings.Contains(name, "uat") {
		if version, exists := labels[constants.KubesphereVersion]; exists && version == constants.KubesphereInitVersion {
			return true
		}
	}
	return false
}
func (d deploymentPredicate) Update(e event.UpdateEvent) bool {
	//if pod label no changes or add labels, ignore
	return false
}
func (d deploymentPredicate) Delete(e event.DeleteEvent) bool {
	return false

}
func (d deploymentPredicate) Generic(e event.GenericEvent) bool {
	return false
}

func (d *DeploymentOperatorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.Deployment{}).
		WithEventFilter(&deploymentPredicate{}).
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
