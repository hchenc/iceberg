package controller

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/hchenc/iceberg/pkg/controllers/filters"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func init() {
	RegisterReconciler("PatchIngress", SetUpIngressReconcile)
}

type IngressOperatorReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func (i *IngressOperatorReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	ingress := &v1.Ingress{}

	err := i.Get(ctx, req.NamespacedName, ingress)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		} else {
			log.Logger.WithFields(logrus.Fields{
				"ingress":   req.Name,
				"namespace": req.Namespace,
				"message":   "failed to reconcile ingress",
			}).Error(err)
			return ctrl.Result{}, err
		}
	}

	if ingress.DeletionTimestamp != nil {
		return ctrl.Result{}, nil
	}

	if uv, exists := ingress.Annotations["nginx.ingress.kubernetes.io/upstream-vhost"]; exists && uv != "" {
		return reconcile.Result{}, nil
	}

	log.Logger.WithFields(logrus.Fields{
		"action": "PatchIngress",
	}).Info("start to action")

	{
		serviceName := ingress.Spec.Rules[0].HTTP.Paths[0].Backend.Service.Name
		namespaceName := ingress.Namespace
		upstreamVhost := fmt.Sprintf("%s.%s.svc.cluster.local", serviceName, namespaceName)
		ingress.Annotations["nginx.ingress.kubernetes.io/upstream-vhost"] = upstreamVhost
		if err := i.Update(ctx, ingress); err != nil {
			log.Logger.WithFields(logrus.Fields{
				"action": "PatchIngress",
			}).Error(err)
			return ctrl.Result{}, err
		}
	}

	log.Logger.WithFields(logrus.Fields{
		"action": "PatchIngress",
	}).Info("finish to action")
	return reconcile.Result{}, nil
}

func (i *IngressOperatorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.Ingress{}).
		WithEventFilter(
			predicate.And(
				&filters.NamespaceCreatePredicate{
					IncludeNamespaces: filters.DefaultIncludeNamespaces,
				},
				&filters.AnnotationCreatePredicate{
					Force: false,
					ExcludeAnnotations: map[string]string{
						"nginx.ingress.kubernetes.io/upstream-vhost":"",
					},
				},
			),
		).
		Complete(i)
}

func SetUpIngressReconcile(mgr manager.Manager) {
	if err := (&IngressOperatorReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("PatchIngress"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		log.Fatalf("unable to create ingress controller for ", err)
	}
}
