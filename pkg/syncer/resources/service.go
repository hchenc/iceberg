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

type serviceInfo struct {
	kubeClient *kubernetes.Clientset
	logger     *logrus.Logger
	ctx        context.Context
}

func (s serviceInfo) Create(obj interface{}) (interface{}, error) {
	service := obj.(*v1.Service)
	svcLogInfo := logrus.Fields{
		"service": service.Name,
	}
	var errs []error
	namespacePrefix := strings.Split(service.Namespace, "-")[0]
	candidates := map[string]string{
		namespacePrefix + "-fat": "fat",
		namespacePrefix + "-uat": "uat",
		namespacePrefix + "-sit": "sit",
	}
	delete(candidates, service.Namespace)

	for namespace := range candidates {
		//service := assembleService(service, namespace)
		service := assembleResource(service, namespace, func(obj interface{}, namespace string) interface{} {
			var newServicePort []v1.ServicePort
			for _, port := range service.Spec.Ports {
				svcPort := v1.ServicePort{
					Name:        port.Name,
					Protocol:    port.Protocol,
					AppProtocol: port.AppProtocol,
					Port:        port.Port,
					TargetPort:  port.TargetPort,
				}
				newServicePort = append(newServicePort, svcPort)
			}
			return &v1.Service{
				TypeMeta: service.TypeMeta,
				ObjectMeta: metav1.ObjectMeta{
					Name:        service.Name,
					Namespace:   namespace,
					Labels:      service.Labels,
					Annotations: service.Annotations,
					Finalizers:  service.Finalizers,
					ClusterName: service.ClusterName,
				},
				Spec: v1.ServiceSpec{
					Ports:           newServicePort,
					Selector:        service.Spec.Selector,
					Type:            service.Spec.Type,
					SessionAffinity: service.Spec.SessionAffinity,
				},
			}
		}).(*v1.Service)
		_, err := s.kubeClient.CoreV1().Services(namespace).Create(s.ctx, service, metav1.CreateOptions{})
		if err == nil || errors.IsAlreadyExists(err) {
			s.logger.WithFields(svcLogInfo).WithFields(logrus.Fields{
				"namespace": namespace,
			}).Info("finish to create namespaced kubernetes service")
		} else {
			s.logger.WithFields(svcLogInfo).WithFields(logrus.Fields{
				"namespace": namespace,
				"message":   "failed to create namespaced kubernetes service",
			}).Error(err)
			errs = append(errs, err)
		}
	}
	if len(errs) != 0 {
		return nil, baseErr.New("failed to sync kubernetes service")
	} else {
		s.logger.WithFields(svcLogInfo).Info("finish to sync kubesphere service")
		return nil, nil
	}
}

func (s serviceInfo) Update(objOld interface{}, objNew interface{}) error {
	panic("implement me")
}

func (s serviceInfo) Delete(name string) error {
	panic("implement me")
}

func (s serviceInfo) GetByName(name string) (interface{}, error) {
	panic("implement me")
}

func (s serviceInfo) GetByID(id int) (interface{}, error) {
	panic("implement me")
}

func (s serviceInfo) List(key string) (interface{}, error) {
	panic("implement me")
}

func NewServiceGenerator(ctx context.Context, kubeClient *kubernetes.Clientset) syncer.Generator {
	logger := utils.GetLogger(logrus.Fields{
		"component": "kubernetes",
		"resource":  "service",
	})
	return serviceInfo{
		kubeClient: kubeClient,
		ctx:        ctx,
		logger:     logger,
	}
}
