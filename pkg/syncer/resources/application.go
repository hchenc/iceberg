package resources

import (
	"context"
	baseErr "errors"
	applicationv1beta1 "github.com/hchenc/application/pkg/apis/app/v1beta1"
	"github.com/hchenc/application/pkg/client/clientset/versioned"
	"github.com/hchenc/iceberg/pkg/syncer"
	"github.com/hchenc/iceberg/pkg/utils"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"strings"
)

type applicationInfo struct {
	appClient  *versioned.Clientset
	kubeClient *kubernetes.Clientset
	logger     *logrus.Logger
	ctx        context.Context
}

func (a applicationInfo) Create(obj interface{}) (interface{}, error) {
	application := obj.(*applicationv1beta1.Application)
	appLogInfo := logrus.Fields{
		"application": application.Name,
	}
	a.logger.WithFields(appLogInfo).Info("start to create kubesphere application")
	namespacePrefix := strings.Split(application.Namespace, "-")[0]
	candidates := map[string]string{
		namespacePrefix + "-fat": "fat",
		namespacePrefix + "-uat": "uat",
		namespacePrefix + "-sit": "sit",
	}
	delete(candidates, application.Namespace)

	var errs []error

	for namespace := range candidates {
		application := assembleResource(application, namespace, func(obj interface{}, namespace string) interface{} {
			return &applicationv1beta1.Application{
				ObjectMeta: v1.ObjectMeta{
					Name:        application.Name,
					Namespace:   namespace,
					Labels:      application.Labels,
					Annotations: application.Annotations,
					Finalizers:  application.Finalizers,
					ClusterName: application.ClusterName,
				},
				Spec: application.Spec,
			}
		}).(*applicationv1beta1.Application)
		_, err := a.appClient.AppV1beta1().Applications(namespace).Create(a.ctx, application, v1.CreateOptions{})
		if err == nil || errors.IsAlreadyExists(err) {
			a.logger.WithFields(appLogInfo).WithFields(logrus.Fields{
				"namespace": namespace,
			}).Info("finish to create namespaced kubesphere application")
		} else {
			a.logger.WithFields(appLogInfo).WithFields(logrus.Fields{
				"namespace": namespace,
				"message":   "failed to create namespaced kubesphere application",
			}).Error(err)
			errs = append(errs, err)
		}
	}
	if len(errs) != 0 {
		return nil, baseErr.New("failed to sync kubesphere application")
	} else {
		a.logger.WithFields(appLogInfo).Info("finish to sync kubesphere application")
		return nil, nil
	}
}

func (a applicationInfo) Update(objOld interface{}, objNew interface{}) error {
	panic("implement me")
}

func (a applicationInfo) Delete(appName string) error {
	panic("implement me")
}

func (a applicationInfo) GetByName(name string) (interface{}, error) {
	panic("implement me")
}

func (a applicationInfo) GetByID(id int) (interface{}, error) {
	panic("implement me")
}

func (a applicationInfo) List(key string) (interface{}, error) {
	panic("implement me")
}

func NewApplicationGenerator(ctx context.Context, kubeClient *kubernetes.Clientset, appClient *versioned.Clientset) syncer.Generator {
	logger := utils.GetLogger(logrus.Fields{
		"component": "kubesphere",
		"resource":  "application",
	})
	return applicationInfo{
		appClient:  appClient,
		kubeClient: kubeClient,
		ctx:        ctx,
		logger:     logger,
	}
}
