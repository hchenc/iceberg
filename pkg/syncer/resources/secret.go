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

type secretInfo struct {
	kubeClient *kubernetes.Clientset
	logger     *logrus.Logger
	ctx        context.Context
}

func (s secretInfo) Create(obj interface{}) (interface{}, error) {
	secret := obj.(*v1.Secret)
	secLogInfo := logrus.Fields{
		"service": secret.Name,
	}
	var errs []error
	namespacePrefix := strings.Split(secret.Namespace, "-")[0]
	candidates := map[string]string{
		namespacePrefix + "-fat": "fat",
		namespacePrefix + "-uat": "uat",
		namespacePrefix + "-sit": "sit",
	}
	delete(candidates, secret.Namespace)

	for namespace := range candidates {
		//service := assembleService(service, namespace)
		itemKey := map[string][]byte{}
		for k := range secret.Data {
			itemKey[k] = []byte("")
		}
		secret := assembleResource(secret, namespace, func(obj interface{}, namespace string) interface{} {
			return &v1.Secret{
				TypeMeta: secret.TypeMeta,
				ObjectMeta: metav1.ObjectMeta{
					Name:        secret.Name,
					Namespace:   namespace,
					Labels:      secret.Labels,
					Annotations: secret.Annotations,
					Finalizers:  secret.Finalizers,
					ClusterName: secret.ClusterName,
				},
				Type: secret.Type,
				Data: itemKey,
			}
		}).(*v1.Secret)
		_, err := s.kubeClient.CoreV1().Secrets(namespace).Create(s.ctx, secret, metav1.CreateOptions{})
		if err == nil || errors.IsAlreadyExists(err) {
			s.logger.WithFields(secLogInfo).WithFields(logrus.Fields{
				"namespace": namespace,
			}).Info("finish to create namespaced kubernetes secret")
		} else {
			s.logger.WithFields(secLogInfo).WithFields(logrus.Fields{
				"namespace": namespace,
				"message":   "failed to create namespaced kubernetes secret",
			}).Error(err)
			errs = append(errs, err)
		}
	}
	if len(errs) != 0 {
		return nil, baseErr.New("failed to sync kubernetes secret")
	} else {
		s.logger.WithFields(secLogInfo).Info("finish to sync kubesphere secret")
		return nil, nil
	}
}

func (s secretInfo) Update(objOld interface{}, objNew interface{}) error {
	panic("implement me")
}

func (s secretInfo) Delete(name string) error {
	panic("implement me")
}

func (s secretInfo) GetByName(name string) (interface{}, error) {
	panic("implement me")
}

func (s secretInfo) GetByID(id int) (interface{}, error) {
	panic("implement me")
}

func (s secretInfo) List(key string) (interface{}, error) {
	panic("implement me")
}

func NewSecretGenerator(ctx context.Context, kubeClient *kubernetes.Clientset) syncer.Generator {
	logger := utils.GetLogger(logrus.Fields{
		"component": "kubernetes",
		"resource":  "service",
	})
	return secretInfo{
		kubeClient: kubeClient,
		ctx:        ctx,
		logger:     logger,
	}
}
