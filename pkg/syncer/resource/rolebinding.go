package resource

import (
	"context"
	baseErr "errors"
	"github.com/hchenc/iceberg/pkg/apis/iam/v1alpha2"
	"github.com/hchenc/iceberg/pkg/constants"
	"github.com/hchenc/iceberg/pkg/syncer"
	"github.com/hchenc/iceberg/pkg/utils"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type rolebindingInfo struct {
	kubeclient *kubernetes.Clientset
	logger     *logrus.Logger
	ctx        context.Context
}

func (r rolebindingInfo) Create(obj interface{}) (interface{}, error) {
	workspaceRolebinding := obj.(*v1alpha2.WorkspaceRoleBinding)
	workspaceName := workspaceRolebinding.Labels[constants.KubesphereWorkspace]
	userName := workspaceRolebinding.Subjects[0].Name

	rbLogInfo := logrus.Fields{
		"rolebinding": workspaceRolebinding.Name,
	}
	var errs []error

	candidates := map[string]string{
		workspaceName + "-fat": "fat",
		workspaceName + "-uat": "uat",
		workspaceName + "-sit": "sit",
	}

	for namespace := range candidates {
		rolebinding := assembleResource(workspaceRolebinding, namespace, func(obj interface{}, namespace string) interface{} {
			return &v1.RoleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name:      userName + "-operator",
					Namespace: namespace,
					Labels: map[string]string{
						"iam.kubesphere.io/user-ref": userName,
					},
				},
				Subjects: []v1.Subject{
					{
						Kind:     "User",
						APIGroup: "rbac.authorization.k8s.io",
						Name:     userName,
					},
				},
				RoleRef: v1.RoleRef{
					APIGroup: "rbac.authorization.k8s.io",
					Kind:     "Role",
					Name:     "operator",
				},
			}
		}).(*v1.RoleBinding)
		_, err := r.kubeclient.RbacV1().RoleBindings(namespace).Create(r.ctx, rolebinding, metav1.CreateOptions{})
		if err == nil || errors.IsAlreadyExists(err) {
			r.logger.WithFields(rbLogInfo).WithFields(logrus.Fields{
				"namespace": namespace,
			}).Info("finish to create namespaced kubesphere rolebinding")
		} else {
			r.logger.WithFields(rbLogInfo).WithFields(logrus.Fields{
				"namespace": namespace,
				"message":   "failed to create namespaced kubesphere rolebinding",
			}).Error(err)
			errs = append(errs, err)
		}
	}
	if len(errs) != 0 {
		return nil, baseErr.New("failed to sync kubesphere rolebinding")
	} else {
		r.logger.WithFields(rbLogInfo).Info("finish to sync kubesphere rolebinding")
		return nil, nil
	}
}

func (r rolebindingInfo) Update(objOld interface{}, objNew interface{}) error {
	panic("implement me")
}

func (r rolebindingInfo) Delete(name string) error {
	panic("implement me")
}

func (r rolebindingInfo) GetByName(name string) (interface{}, error) {
	panic("implement me")
}

func (r rolebindingInfo) GetByID(id int) (interface{}, error) {
	panic("implement me")
}

func (r rolebindingInfo) List(key string) (interface{}, error) {
	panic("implement me")
}

func NewRolebindingGenerator(ctx context.Context, kubeclient *kubernetes.Clientset) syncer.Generator {
	logger := utils.GetLogger(logrus.Fields{
		"component": "kubesphere",
		"resource":  "rolebinding",
	})
	return rolebindingInfo{
		kubeclient: kubeclient,
		ctx:        ctx,
		logger:     logger,
	}
}
