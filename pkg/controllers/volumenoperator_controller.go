package controller

import (
	"context"
	"github.com/go-logr/logr"
	"github.com/hchenc/iceberg/pkg/constants"
	"github.com/hchenc/iceberg/pkg/controllers/filters"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"strings"
	"time"
)

func init() {
	RegisterReconciler("PersistentVolume", SetUpVolumeReconcile)
}

type VolumeOperatorReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func (v VolumeOperatorReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	volume := &v1.PersistentVolumeClaim{}

	err := v.Get(ctx, req.NamespacedName, volume)
	if err != nil {
		if errors.IsNotFound(err) {
			v.Log.Info("receive delete event")
		} else {
			log.Logger.WithFields(logrus.Fields{
				"pvc":     req.NamespacedName,
				"message": "failed to reconcile pvc",
			}).Error(err)
		}
	} else {
		log.Logger.WithFields(logrus.Fields{
			"action": "PersistentVolume",
		}).Info("start to action")
		//sync volume to all environment(fat|uat|sit)
		_, err := volumeGeneratorService.Add(volume)
		if err != nil {
			log.Logger.WithFields(logrus.Fields{
				"event":    "create",
				"resource": "Volume",
				"name":     volume.Name,
				"result":   "failed",
				"error":    err.Error(),
			}).Errorf("volume created failed, retry after %d second", RetryPeriod)
			return reconcile.Result{
				RequeueAfter: RetryPeriod * time.Second,
			}, err
		}
		log.Logger.WithFields(logrus.Fields{
			"event":    "create",
			"resource": "Volume",
			"name":     volume.Name,
			"result":   "success",
			"message":  "volume controller successful",
		}).Infof("finish to sync volume %s", volume.Name)
	}
	log.Logger.WithFields(logrus.Fields{
		"action": "PersistentVolume",
	}).Info("finish to action")
	return reconcile.Result{}, nil
}

func (v *VolumeOperatorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.PersistentVolumeClaim{}).
		WithEventFilter(
			predicate.And(
				&filters.NamespaceCreatePredicate{
					IncludeNamespaces: filters.DefaultIncludeNamespaces,
				},
				&filters.LabelCreatePredicate{
					Force: false,
					IncludeLabels: map[string]string{
						constants.KubesphereVersion: constants.KubesphereInitVersion,
					}},
			),
		).
		Complete(v)
}

type volumePredicate struct {
}

func (v volumePredicate) Create(e event.CreateEvent) bool {
	namespace := e.Object.GetNamespace()
	if _, exist := e.Object.GetLabels()[constants.KubesphereAppName]; !exist {
		return false
	} else if strings.Contains(namespace, "sit") || strings.Contains(namespace, "fat") || strings.Contains(namespace, "uat") {
		return true
	} else {
		return false
	}
}

func (v volumePredicate) Delete(event.DeleteEvent) bool {
	return false
}

func (v volumePredicate) Update(event.UpdateEvent) bool {
	return false
}

func (v volumePredicate) Generic(event.GenericEvent) bool {
	return false
}

func SetUpVolumeReconcile(mgr manager.Manager) {
	if err := (&VolumeOperatorReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("PersistentVolume"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		log.Fatalf("unable to create volume controller for", err)
	}
}
