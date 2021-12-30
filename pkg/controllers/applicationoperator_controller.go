package controller

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"strings"
	"time"

	"github.com/hchenc/application/pkg/apis/app/v1beta1"
)

func init() {
	RegisterReconciler("AppToProject", SetUpProjectReconcile)
}

type ApplicationOperatorReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func (r *ApplicationOperatorReconciler) Reconcile(req reconcile.Request) (reconcile.Result, error) {
	application := &v1beta1.Application{}
	ctx := context.Background()

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

type projectPredicate struct {
}

func (r projectPredicate) Create(e event.CreateEvent) bool {
	name := e.Meta.GetNamespace()
	if strings.Contains(name, "system") || strings.Contains(name, "kube") {
		return false
	} else if strings.Contains(name, "sit") || strings.Contains(name, "fat") || strings.Contains(name, "uat") {
		return true
	} else {
		return false
	}
}
func (r projectPredicate) Update(e event.UpdateEvent) bool {
	//if pod label no changes or add labels, ignore
	return false
}
func (r projectPredicate) Delete(e event.DeleteEvent) bool {
	name := e.Meta.GetName()
	if strings.Contains(name, "system") || strings.Contains(name, "kube") {
		return false
	} else {
		return true
	}
}
func (r projectPredicate) Generic(e event.GenericEvent) bool {
	return false
}

func (r *ApplicationOperatorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.Application{}).
		WithEventFilter(&projectPredicate{}).
		Complete(r)
}

func SetUpProjectReconcile(mgr manager.Manager) {
	if err := (&ApplicationOperatorReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("AppToProject"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		log.Fatalf("unable to create application controller for ", err)
	}
}
