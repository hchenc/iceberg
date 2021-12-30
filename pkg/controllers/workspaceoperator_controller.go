package controller

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/hchenc/iceberg/pkg/apis/tenant/v1alpha2"
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
)

var (
	action string = "WorkspaceTemplateToGroup"
)

func init() {
	RegisterReconciler(action, SetUpGroupReconcile)
}

type WorkspaceOperatorReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func (g *WorkspaceOperatorReconciler) Reconcile(req reconcile.Request) (reconcile.Result, error) {
	workspaceTemplate := &v1alpha2.WorkspaceTemplate{}
	ctx := context.Background()

	err := g.Get(ctx, req.NamespacedName, workspaceTemplate)
	if err != nil {
		if errors.IsNotFound(err) {
			err := groupGeneratorService.Delete(req.Name)
			if err != nil {
				log.Logger.WithFields(logrus.Fields{
					"workspaceTemplate": req.Name,
					"message":           "failed to delete workspaceTemplate",
				}).Error(err)
			}
		} else {
			log.Logger.WithFields(logrus.Fields{
				"workspaceTemplate": req.Name,
				"message":           "failed to reconcile workspaceTemplate",
			}).Error(err)
		}
	} else {
		log.Logger.WithFields(logrus.Fields{
			"action": action,
		}).Info("start to action")
		// create gitlab group
		gitlabGroup, err := groupGeneratorService.Add(workspaceTemplate)
		if err != nil {
			if gitlabGroup != nil {
				log.Logger.WithFields(logrus.Fields{
					"event":    "create",
					"resource": "Pager",
					"name":     "workspace-" + workspaceTemplate.Name,
					"result":   "failed",
					"error":    err.Error(),
				}).Errorf("pager created failed, retry after %d second", RetryPeriod)
			} else {
				log.Logger.WithFields(logrus.Fields{
					"event":    "create",
					"resource": "Group",
					"name":     workspaceTemplate.Name,
					"result":   "failed",
					"error":    err.Error(),
				}).Errorf("group created failed, retry after %d second", RetryPeriod)
			}
			return reconcile.Result{
				RequeueAfter: RetryPeriod * time.Second,
			}, err
		}

		// create KubeSphere's project(namespace) as environment
		_, err = namespaceGeneratorService.Add(workspaceTemplate)
		if err != nil {
			log.Logger.WithFields(logrus.Fields{
				"event":    "create",
				"resource": "Namespace",
				"name":     workspaceTemplate.Name,
				"result":   "failed",
				"error":    err.Error(),
			}).Errorf("namespace created failed, retry after %d second", RetryPeriod)
			return reconcile.Result{
				RequeueAfter: RetryPeriod * time.Second,
			}, err
		}

		// create harbor's project
		_, err = harborGeneratorService.Add(workspaceTemplate)
		if err != nil {
			log.Logger.WithFields(logrus.Fields{
				"event":    "create",
				"resource": "Harbor",
				"name":     workspaceTemplate.Name,
				"result":   "failed",
				"error":    err.Error(),
			}).Errorf("harbor project created failed, retry after %d second", RetryPeriod)
			return reconcile.Result{
				RequeueAfter: RetryPeriod * time.Second,
			}, err
		}
		log.Logger.WithFields(logrus.Fields{
			"event":    "create",
			"resource": "Workspace",
			"name":     workspaceTemplate.Name,
			"result":   "success",
		}).Infof("finish to sync workspace %s", workspaceTemplate.Name)
	}
	log.Logger.WithFields(logrus.Fields{
		"action": action,
	}).Info("finish to action")
	return reconcile.Result{}, nil
}

func (g *WorkspaceOperatorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha2.WorkspaceTemplate{}).
		WithEventFilter(&workspacePredicate{}).
		Complete(g)
}

type workspacePredicate struct {
}

func (r workspacePredicate) Create(e event.CreateEvent) bool {
	name := e.Meta.GetName()
	if strings.Contains(name, "system") || strings.Contains(name, "kube") {
		return false
	} else {
		return true
	}
}
func (r workspacePredicate) Update(e event.UpdateEvent) bool {
	//if pod label no changes or add labels, ignore
	return false
}
func (r workspacePredicate) Delete(e event.DeleteEvent) bool {
	name := e.Meta.GetName()
	if strings.Contains(name, "system") || strings.Contains(name, "kube") {
		return false
	} else {
		return true
	}
}
func (r workspacePredicate) Generic(e event.GenericEvent) bool {
	return false
}

func SetUpGroupReconcile(mgr manager.Manager) {
	if err := (&WorkspaceOperatorReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName(action),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		log.Fatalf("unable to create workspace controller for", err)
	}
}
