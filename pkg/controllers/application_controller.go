package controller

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/hchenc/iceberg/pkg/controllers/filters"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"

	"github.com/hchenc/application/pkg/apis/app/v1beta1"
)

func init() {
	RegisterReconciler("AppToProject", SetUpProjectReconciler)
}

type ApplicationReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func (r *ApplicationReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	application := &v1beta1.Application{}

	err := r.Get(ctx, req.NamespacedName, application)
	if err != nil {
		if errors.IsNotFound(err) {
			err := projectGeneratorService.Delete(req.Name)
			if err != nil {
				log.Logger.WithFields(logrus.Fields{
					"application": req.Name,
					"namespace":   req.Namespace,
					"message":     "failed to delete application",
				}).Error(err)
			}
		} else {
			log.Logger.WithFields(logrus.Fields{
				"application": req.Name,
				"namespace":   req.Namespace,
				"message":     "failed to reconcile application",
			}).Error(err)
		}
	} else {
		log.Logger.WithFields(logrus.Fields{
			"action": "AppToProject",
		}).Info("start to action")
		// create gitlab project
		project, err := projectGeneratorService.Add(application)
		if err != nil {
			if project != nil {
				log.Logger.WithFields(logrus.Fields{
					"event":    "create",
					"resource": "Pager",
					"name":     "application-" + application.Name,
					"result":   "failed",
					"error":    err.Error(),
				}).Errorf("pager created failed, retry after %d second", RetryPeriod)
			} else {
				log.Logger.WithFields(logrus.Fields{
					"event":    "create",
					"resource": "Project",
					"name":     application.Name,
					"result":   "failed",
					"error":    err.Error(),
				}).Errorf("project created failed, retry after %d second", RetryPeriod)
			}
			return reconcile.Result{
				RequeueAfter: RetryPeriod * time.Second,
			}, err
		}

		//sync application to all environment(fat|uat|sit)
		_, err = applicationGeneratorService.Add(application)
		if err != nil {
			log.Logger.WithFields(logrus.Fields{
				"event":    "create",
				"resource": "Application",
				"name":     application.Name,
				"result":   "failed",
				"error":    err.Error(),
			}).Errorf("application created failed, retry after %d second", RetryPeriod)
			return reconcile.Result{
				RequeueAfter: RetryPeriod * time.Second,
			}, err
		}

		log.Logger.WithFields(logrus.Fields{
			"event":    "create",
			"resource": "Application",
			"name":     application.Name,
			"result":   "success",
		}).Infof("finish to sync application %s", application.Name)
	}
	log.Logger.WithFields(logrus.Fields{
		"action": "AppToProject",
	}).Info("finish to action")
	return reconcile.Result{}, nil
}

func (r *ApplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.Application{}).
		WithEventFilter(
			predicate.Or(
				&filters.NamespaceCreatePredicate{
					IncludeNamespaces: filters.DefaultIncludeNamespaces,
				},
				&filters.NamespaceDeletePredicate{
					ExcludeNamespaces: filters.DefaultExcludeNamespaces,
				},
			),
		).
		Complete(r)
}

func SetUpProjectReconciler(mgr manager.Manager) {
	if err := (&ApplicationReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("AppToProject"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		log.Fatalf("unable to create application controller for ", err)
	}
}
