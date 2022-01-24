package controller

import (
	"context"
	"github.com/go-logr/logr"
	iamv1alpha2 "github.com/hchenc/iceberg/pkg/apis/iam/v1alpha2"
	"github.com/hchenc/iceberg/pkg/controllers/filters"
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

func init() {
	RegisterReconciler("UserToUser", SetUpUserReconcile)
}

type UserOperatorReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func (u *UserOperatorReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	user := &iamv1alpha2.User{}

	err := u.Get(ctx, req.NamespacedName, user)
	if err != nil {
		if errors.IsNotFound(err) {
			err := userGeneratorService.Delete(req.Name)
			if err != nil {
				log.Logger.WithFields(logrus.Fields{
					"user":      req.Name,
					"namespace": req.Namespace,
					"message":   "failed to delete user",
				}).Error(err)
			}
		} else {
			log.Logger.WithFields(logrus.Fields{
				"user":      req.Name,
				"namespace": req.Namespace,
				"message":   "failed to reconcile user",
			}).Error(err)
		}
	} else {
		log.Logger.WithFields(logrus.Fields{
			"action": "UserToUser",
		}).Info("start to action")
		// create gitlab project
		gitlabUser, err := userGeneratorService.Add(user)
		if err != nil {
			if gitlabUser != nil {
				log.Logger.WithFields(logrus.Fields{
					"event":    "create",
					"resource": "Pager",
					"name":     "user-" + user.Name,
					"result":   "failed",
					"error":    err.Error(),
				}).Errorf("pager created failed, retry after %d second", RetryPeriod)
			} else {
				log.Logger.WithFields(logrus.Fields{
					"event":    "create",
					"resource": "User",
					"name":     user.Name,
					"result":   "failed",
					"error":    err.Error(),
				}).Errorf("user created failed, retry after %d second", RetryPeriod)
			}
			return reconcile.Result{
				RequeueAfter: RetryPeriod * time.Second,
			}, err
		}
		log.Logger.WithFields(logrus.Fields{
			"event":    "create",
			"resource": "User",
			"name":     user.Name,
			"result":   "success",
		}).Infof("finish to sync user %s", user.Name)
	}
	log.Logger.WithFields(logrus.Fields{
		"action": "UserToUser",
	}).Info("finish to action")
	return reconcile.Result{}, nil
}

type userPredicate struct {
}

func (r userPredicate) Create(e event.CreateEvent) bool {
	name := e.Object.GetName()
	if strings.Contains(name, "system") || strings.Contains(name, "admin") {
		return false
	} else {
		return true
	}
}
func (r userPredicate) Update(e event.UpdateEvent) bool {
	//if pod label no changes or add labels, ignore
	return false
}
func (r userPredicate) Delete(e event.DeleteEvent) bool {
	return false

}
func (r userPredicate) Generic(e event.GenericEvent) bool {
	return false
}

func (u *UserOperatorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&iamv1alpha2.User{}).
		WithEventFilter(&filters.NameCreatePredicate{
			ExcludeNames: filters.DefaultExcludeNames,
		}).
		Complete(u)
}

func SetUpUserReconcile(mgr manager.Manager) {
	if err := (&UserOperatorReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("UserToUser"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		log.Fatalf("unable to create user controller for", err)
	}
}
