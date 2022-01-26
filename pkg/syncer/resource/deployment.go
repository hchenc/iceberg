package resource

import (
	"context"
	baseErr "errors"
	"fmt"
	"github.com/hchenc/iceberg/pkg/constants"
	"github.com/hchenc/iceberg/pkg/syncer"
	"github.com/hchenc/iceberg/pkg/utils"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"strings"
)

type deploymentInfo struct {
	kubeClient *kubernetes.Clientset
	logger     *logrus.Logger
	ctx        context.Context
}

func (d deploymentInfo) Create(obj interface{}) (interface{}, error) {
	deployment := obj.(*v1.Deployment)
	dpLogInfo := logrus.Fields{
		"deployment": deployment.Name,
	}
	var errs []error
	namespacePrefix := strings.Split(deployment.Namespace, "-")[0]
	candidates := map[string]string{
		namespacePrefix + "-fat": "fat",
		namespacePrefix + "-uat": "uat",
		namespacePrefix + "-sit": "sit",
	}
	delete(candidates, deployment.Namespace)

	for namespace := range candidates {
		deployment := assembleResource(deployment, namespace, func(obj interface{}, namespace string) interface{} {
			return &v1.Deployment{
				TypeMeta: deployment.TypeMeta,
				ObjectMeta: metav1.ObjectMeta{
					Name:        deployment.Name,
					Namespace:   namespace,
					Labels:      deployment.Labels,
					Annotations: deployment.Annotations,
					Finalizers:  deployment.Finalizers,
					ClusterName: deployment.ClusterName,
				},
				Spec: v1.DeploymentSpec{
					Replicas: deployment.Spec.Replicas,
					Selector: deployment.Spec.Selector,
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels:      deployment.Spec.Template.Labels,
							Annotations: deployment.Spec.Template.Annotations,
						},
						Spec: corev1.PodSpec{
							Containers:         deployment.Spec.Template.Spec.Containers,
							ServiceAccountName: deployment.Spec.Template.Spec.ServiceAccountName,
							Affinity:           deployment.Spec.Template.Spec.Affinity,
							InitContainers:     deployment.Spec.Template.Spec.InitContainers,
							Volumes:            deployment.Spec.Template.Spec.Volumes,
							ImagePullSecrets:   deployment.Spec.Template.Spec.ImagePullSecrets,
						},
					},
					Strategy: deployment.Spec.Strategy,
				},
			}
		}).(*v1.Deployment)
		if dps, err := d.kubeClient.AppsV1().Deployments(deployment.Namespace).List(context.Background(), metav1.ListOptions{
			LabelSelector: fmt.Sprintf("%s=%s", constants.KubesphereAppName, deployment.Labels[constants.KubesphereAppName]),
		}); err == nil {
			if len(dps.Items) != 0 {
				return nil, nil
			}
		} else {
			return nil, err
		}
		_, err := d.kubeClient.AppsV1().Deployments(namespace).Create(d.ctx, deployment, metav1.CreateOptions{})
		if err == nil || errors.IsAlreadyExists(err) {
			d.logger.WithFields(dpLogInfo).WithFields(logrus.Fields{
				"namespace": namespace,
			}).Info("finish to create namespaced kubernetes deployment")
		} else {
			d.logger.WithFields(dpLogInfo).WithFields(logrus.Fields{
				"namespace": namespace,
				"message":   "failed to create namespaced kubernetes deployment",
			}).Error(err)
			errs = append(errs, err)
		}
	}
	if len(errs) != 0 {
		return nil, baseErr.New("failed to sync kubernetes deployment")
	} else {
		d.logger.WithFields(dpLogInfo).Info("finish to sync kubernetes deployment")
		return nil, nil
	}
}

func (d deploymentInfo) Update(objOld interface{}, objNew interface{}) error {
	panic("implement me")
}

func (d deploymentInfo) Delete(name string) error {
	panic("implement me")
}

func (d deploymentInfo) GetByName(name string) (interface{}, error) {
	panic("implement me")
}

func (d deploymentInfo) GetByID(id int) (interface{}, error) {
	panic("implement me")
}

func (d deploymentInfo) List(key string) (interface{}, error) {
	panic("implement me")
}

func NewDeploymentGenerator(ctx context.Context, kubeClient *kubernetes.Clientset) syncer.Generator {
	logger := utils.GetLogger(logrus.Fields{
		"component": "kubernetes",
		"resource":  "deployment",
	})
	return deploymentInfo{
		kubeClient: kubeClient,
		ctx:        ctx,
		logger:     logger,
	}
}
