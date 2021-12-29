package resource

import (
	"context"
	baseErr "errors"
	tenantv1alpha1 "github.com/hchenc/iceberg/pkg/apis/tenant/v1alpha1"
	"github.com/hchenc/iceberg/pkg/apis/tenant/v1alpha2"
	"github.com/hchenc/iceberg/pkg/constants"
	"github.com/hchenc/iceberg/pkg/syncer"
	"github.com/hchenc/iceberg/pkg/utils"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	env = map[string]string{
		"fat": constants.FAT,
		"sit": constants.SIT,
		"uat": constants.UAT,
	}
)

type namespaceInfo struct {
	client *kubernetes.Clientset
	logger *logrus.Logger
	ctx    context.Context
}

func (n namespaceInfo) Create(obj interface{}) (interface{}, error) {
	workspace := obj.(*v1alpha2.WorkspaceTemplate)
	workspaceName := workspace.Name
	nsLogInfo := logrus.Fields{
		"workspace": workspaceName,
	}
	var errs []error
	candidates := map[string]string{
		"fat": workspaceName + "-fat",
		"uat": workspaceName + "-uat",
		"sit": workspaceName + "-sit",
	}
	creator := workspace.GetAnnotations()[constants.KubesphereCreator]

	for index, namespaceName := range candidates {
		namespace := assembleResource(workspace, namespaceName, func(obj interface{}, namespace string) interface{} {
			return &v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: namespaceName,
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion: tenantv1alpha1.SchemeGroupVersion.Group,
							Kind:       "Workspace",
							Name:       workspaceName,
							UID:        workspace.UID,
						},
					},
					Labels: map[string]string{
						"kubesphere.io/creator":       creator,
						"kubernetes.io/metadata.name": namespaceName,
						"kubesphere.io/namespace":     namespaceName,
						"kubesphere.io/workspace":     workspaceName,
					},
					Annotations: map[string]string{
						"kubesphere.io/creator":     creator,
						"kubesphere.io/description": env[index],
					},
				},
			}
		}).(*v1.Namespace)
		_, err := n.client.CoreV1().Namespaces().Create(n.ctx, namespace, metav1.CreateOptions{})
		if err == nil || errors.IsAlreadyExists(err) {
			n.logger.WithFields(nsLogInfo).WithFields(logrus.Fields{
				"namespace": namespace.Name,
			}).Info("finish to create namespaced kubernetes namespace")
		} else {
			n.logger.WithFields(nsLogInfo).WithFields(logrus.Fields{
				"namespace": namespace.Name,
				"message":   "failed to create namespaced kubernetes namespace",
			}).Error(err)
			errs = append(errs, err)
		}
	}
	if len(errs) != 0 {
		return nil, baseErr.New("failed to sync kubernetes namespace")
	} else {
		n.logger.WithFields(nsLogInfo).Info("finish to sync kubernetes namespace")
		return nil, nil
	}
}

func (n namespaceInfo) Update(objOld interface{}, objNew interface{}) error {
	panic("implement me")
}

func (n namespaceInfo) Delete(name string) error {
	panic("implement me")
}

func (n namespaceInfo) GetByName(key string) (interface{}, error) {
	ctx := context.Background()

	ns, err := n.client.CoreV1().Namespaces().Get(ctx, key, metav1.GetOptions{})
	return ns, err
}

func (n namespaceInfo) GetByID(key int) (interface{}, error) {
	panic("implement me")
}

func (n namespaceInfo) List(key string) (interface{}, error) {
	panic("implement me")
}

func NewNamespaceGenerator(ctx context.Context, client *kubernetes.Clientset) syncer.Generator {
	logger := utils.GetLogger(logrus.Fields{
		"component": "kubernetes",
		"resource":  "namespace",
	})
	return &namespaceInfo{
		client: client,
		ctx:    ctx,
		logger: logger,
	}
}
