package gitlab

import (
	"context"
	iamv1alpha2 "github.com/hchenc/iceberg/pkg/apis/iam/v1alpha2"
	"github.com/hchenc/iceberg/pkg/clients/clientset"
	"github.com/hchenc/iceberg/pkg/constants"
	"github.com/hchenc/iceberg/pkg/syncer"
	"github.com/hchenc/iceberg/pkg/utils"
	utilerrors "github.com/hchenc/iceberg/pkg/utils/errors"
	"github.com/hchenc/pager/pkg/apis/devops/v1alpha1"
	pager "github.com/hchenc/pager/pkg/client/clientset/versioned"
	"github.com/sirupsen/logrus"
	git "github.com/xanzy/go-gitlab"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"
	"strings"
)

type memberInfo struct {
	gitlabClient *clientset.GitlabClient
	pagerClient  *pager.Clientset
	logger       *logrus.Logger
	ctx          context.Context
}

func (m memberInfo) Create(obj interface{}) (interface{}, error) {
	rolebinding := obj.(*iamv1alpha2.WorkspaceRoleBinding)
	groupName := rolebinding.Labels[constants.KubesphereWorkspace]
	userName := rolebinding.Subjects[0].Name
	if exist, err := m.pagerClient.DevopsV1alpha1().Pagers(constants.DevopsNamespace).Get(m.ctx, "member-"+userName, v1.GetOptions{}); err == nil && exist != nil {
		return nil, nil
	}
	groupRecord, _ := m.pagerClient.DevopsV1alpha1().Pagers(constants.DevopsNamespace).Get(m.ctx, "workspace-"+groupName, v1.GetOptions{})

	userRecord, _ := m.pagerClient.DevopsV1alpha1().Pagers(constants.DevopsNamespace).Get(m.ctx, "user-"+userName, v1.GetOptions{})

	uid, _ := strconv.Atoi(userRecord.Spec.MessageID)

	member, resp, err := m.gitlabClient.Client.GroupMembers.AddGroupMember(groupRecord.Spec.MessageID, &git.AddGroupMemberOptions{
		UserID:      git.Int(uid),
		AccessLevel: git.AccessLevel(git.DeveloperPermissions),
		ExpiresAt:   nil,
	})
	defer resp.Body.Close()

	if err := utilerrors.NewConflict(err); err == nil || errors.IsConflict(err) {
		if member == nil {
			var err error
			member, resp, err = m.gitlabClient.Client.GroupMembers.GetGroupMember(groupRecord.Spec.MessageID, uid)
			defer resp.Body.Close()
			if err != nil {
				return nil, err
			}
		}
		_, err := m.pagerClient.
			DevopsV1alpha1().
			Pagers(constants.DevopsNamespace).
			Create(m.ctx, &v1alpha1.Pager{
				ObjectMeta: v1.ObjectMeta{
					Name: "member-" + member.Username,
				},
				Spec: v1alpha1.PagerSpec{
					MessageID:   strconv.Itoa(member.ID),
					MessageName: member.Username,
					MessageType: rolebinding.Kind,
				},
			}, v1.CreateOptions{})
		if err == nil || errors.IsAlreadyExists(err) {
			return member, nil
		} else {
			return member, err
		}
	}
	return nil, err
}

func (m memberInfo) Update(objOld interface{}, objNew interface{}) error {
	panic("implement me")
}

func (m memberInfo) Delete(rolebindingName string) error {
	pagerName := "member-" + strings.Split(rolebindingName, "-")[0]
	memberLogInfo := logrus.Fields{
		"rolebinding": rolebindingName,
	}
	m.logger.WithFields(memberLogInfo).Info("start to delete gitlab member pager")

	err := m.pagerClient.DevopsV1alpha1().Pagers(constants.DevopsNamespace).Delete(m.ctx, pagerName, v1.DeleteOptions{})
	if err == nil || errors.IsNotFound(err) {
		m.logger.WithFields(memberLogInfo).WithFields(logrus.Fields{
			"pager": pagerName,
		}).Info("finish to delete gitlab member pager")
		return nil
	} else {
		m.logger.WithFields(memberLogInfo).WithFields(logrus.Fields{
			"message": "failed to delete gitlab member pager",
			"pager":   pagerName,
		}).Error(err)
		return err
	}
}

func (m memberInfo) GetByName(name string) (interface{}, error) {
	panic("implement me")
}

func (m memberInfo) GetByID(id int) (interface{}, error) {
	panic("implement me")
}

func (m memberInfo) List(key string) (interface{}, error) {
	panic("implement me")
}

func (m memberInfo) list(key string) (interface{}, error) {
	panic("implement me")
}

func NewMemberGenerator(ctx context.Context, gitlabClient *clientset.GitlabClient, pagerClient *pager.Clientset) syncer.Generator {
	logger := utils.GetLogger(logrus.Fields{
		"component": "gitlab",
		"resource":  "member",
	})
	return &memberInfo{
		gitlabClient: gitlabClient,
		pagerClient:  pagerClient,
		logger:       logger,
		ctx:          ctx,
	}
}
