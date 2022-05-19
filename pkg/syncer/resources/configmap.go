package resources

import (
	"context"
	baseErr "errors"
	"github.com/hchenc/iceberg/pkg/syncer"
	"github.com/hchenc/iceberg/pkg/utils"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"strings"
)

type configmapInfo struct {
	kubeClient *kubernetes.Clientset
	logger     *logrus.Logger
	ctx        context.Context
}

func (c configmapInfo) Create(obj interface{}) (interface{}, error) {
	configmap := obj.(*v1.ConfigMap)
	configmapInfo := logrus.Fields{
		"configmap": configmap.Name,
	}
	var errs []error
	namespacePrefix := strings.Split(configmap.Namespace, "-")[0]
	candidates := map[string]string{
		namespacePrefix + "-fat": "fat",
		namespacePrefix + "-uat": "uat",
		namespacePrefix + "-sit": "sit",
	}
	delete(candidates, configmap.Namespace)

	for namespace := range candidates {
		//service := assembleService(service, namespace)
		if exist, err := c.kubeClient.CoreV1().ConfigMaps(namespace).Get(c.ctx, configmap.Name, metav1.GetOptions{}); err == nil && exist != nil {
			continue
		}
		itemKey := map[string]string{}
		for k := range configmap.Data {
			itemKey[k] = ""
		}
		configmap := assembleResource(configmap, namespace, func(obj interface{}, namespace string) interface{} {
			return &v1.ConfigMap{
				TypeMeta: configmap.TypeMeta,
				ObjectMeta: metav1.ObjectMeta{
					Name:        configmap.Name,
					Namespace:   namespace,
					Labels:      configmap.Labels,
					Annotations: configmap.Annotations,
					Finalizers:  configmap.Finalizers,
					ClusterName: configmap.ClusterName,
				},
				Data: itemKey,
			}
		}).(*v1.ConfigMap)
		_, err := c.kubeClient.CoreV1().ConfigMaps(namespace).Create(c.ctx, configmap, metav1.CreateOptions{})
		if err == nil || errors.IsAlreadyExists(err) {
			c.logger.WithFields(configmapInfo).WithFields(logrus.Fields{
				"namespace": namespace,
			}).Info("finish to create namespaced kubernetes configmap")
		} else {
			c.logger.WithFields(configmapInfo).WithFields(logrus.Fields{
				"namespace": namespace,
				"message":   "failed to create namespaced kubernetes configmap",
			}).Error(err)
			errs = append(errs, err)
		}
	}
	if len(errs) != 0 {
		return nil, baseErr.New("failed to sync kubernetes configmap")
	} else {
		c.logger.WithFields(configmapInfo).Info("finish to sync kubesphere configmap")
		return nil, nil
	}
}

func (c configmapInfo) Update(objOld interface{}, objNew interface{}) error {
	panic("implement me")
}

func (c configmapInfo) Delete(name string) error {
	panic("implement me")
}

func (c configmapInfo) GetByName(name string) (interface{}, error) {
	panic("implement me")
}

func (c configmapInfo) GetByID(id int) (interface{}, error) {
	panic("implement me")
}

func (c configmapInfo) List(key string) (interface{}, error) {
	panic("implement me")
}

func NewConfigMapGenerator(ctx context.Context, kubeClient *kubernetes.Clientset) syncer.Generator {
	logger := utils.GetLogger(logrus.Fields{
		"component": "kubernetes",
		"resource":  "configmap",
	})
	return configmapInfo{
		kubeClient: kubeClient,
		ctx:        ctx,
		logger:     logger,
	}
}
