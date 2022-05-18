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

type volumeInfo struct {
	kubeClient *kubernetes.Clientset
	logger     *logrus.Logger
	ctx        context.Context
}

func (v volumeInfo) Create(obj interface{}) (interface{}, error) {
	volume := obj.(*v1.PersistentVolumeClaim)
	volumeLogInfo := logrus.Fields{
		"volume": volume.Name,
	}
	var errs []error
	namespacePrefix := strings.Split(volume.Namespace, "-")[0]
	candidates := map[string]string{
		namespacePrefix + "-fat": "fat",
		namespacePrefix + "-uat": "uat",
		namespacePrefix + "-sit": "sit",
	}
	delete(candidates, volume.Namespace)

	for namespace := range candidates {
		volume := assembleResource(volume, namespace, func(obj interface{}, namespace string) interface{} {
			return &v1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:        volume.Name,
					Namespace:   namespace,
					Labels:      volume.Labels,
					Annotations: volume.Annotations,
				},
				Spec: v1.PersistentVolumeClaimSpec{
					AccessModes:      volume.Spec.AccessModes,
					Resources:        volume.Spec.Resources,
					StorageClassName: volume.Spec.StorageClassName,
				},
			}
		}).(*v1.PersistentVolumeClaim)
		_, err := v.kubeClient.CoreV1().PersistentVolumeClaims(namespace).Create(v.ctx, volume, metav1.CreateOptions{})
		if err == nil || errors.IsAlreadyExists(err) {
			v.logger.WithFields(volumeLogInfo).WithFields(logrus.Fields{
				"namespace": namespace,
			}).Info("finish to create namespaced kubernetes volume")
		} else {
			v.logger.WithFields(volumeLogInfo).WithFields(logrus.Fields{
				"namespace": namespace,
				"message":   "failed to create namespaced kubernetes volume",
			}).Error(err)
			errs = append(errs, err)
		}
	}
	if len(errs) != 0 {
		return nil, baseErr.New("failed to sync kubernetes volume")
	} else {
		v.logger.WithFields(volumeLogInfo).Info("finish to sync kubesphere volume")
		return nil, nil
	}
}

func (v volumeInfo) Update(objOld interface{}, objNew interface{}) error {
	panic("implement me")
}

func (v volumeInfo) Delete(name string) error {
	panic("implement me")
}

func (v volumeInfo) GetByName(name string) (interface{}, error) {
	panic("implement me")
}

func (v volumeInfo) GetByID(id int) (interface{}, error) {
	panic("implement me")
}

func (v volumeInfo) List(key string) (interface{}, error) {
	panic("implement me")
}

func NewVolumeGenerator(ctx context.Context, clientset *kubernetes.Clientset) syncer.Generator {
	logger := utils.GetLogger(logrus.Fields{
		"component": "kubernetes",
		"resource":  "volume",
	})
	return volumeInfo{
		kubeClient: clientset,
		logger:     logger,
		ctx:        ctx,
	}
}
