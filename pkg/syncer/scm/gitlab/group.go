package gitlab

import (
	"context"
	"github.com/hchenc/iceberg/pkg/apis/tenant/v1alpha2"
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
)

type groupInfo struct {
	gitlabClient *clientset.GitlabClient
	pagerClient  *pager.Clientset
	logger       *logrus.Logger
	ctx          context.Context
	groupName    string
}

func (g groupInfo) Create(obj interface{}) (interface{}, error) {
	workspace := obj.(*v1alpha2.WorkspaceTemplate)
	workspaceLogInfo := logrus.Fields{
		"workspace": workspace.Name,
	}
	g.logger.WithFields(workspaceLogInfo).Info("start to create gitlab group")

	name := git.String(workspace.Name)
	description := git.String(workspace.GetAnnotations()[constants.KubesphereDescription])
	group, resp, err := g.gitlabClient.Client.Groups.CreateGroup(&git.CreateGroupOptions{
		Name:                           name,
		Path:                           name,
		Description:                    description,
		MembershipLock:                 git.Bool(false),
		Visibility:                     git.Visibility(git.PrivateVisibility),
		ShareWithGroupLock:             git.Bool(false),
		RequireTwoFactorAuth:           git.Bool(false),
		TwoFactorGracePeriod:           nil,
		ProjectCreationLevel:           git.ProjectCreationLevel(git.DeveloperProjectCreation),
		AutoDevopsEnabled:              git.Bool(false),
		SubGroupCreationLevel:          git.SubGroupCreationLevel(git.MaintainerSubGroupCreationLevelValue),
		EmailsDisabled:                 git.Bool(false),
		MentionsDisabled:               git.Bool(false),
		LFSEnabled:                     nil,
		RequestAccessEnabled:           nil,
		ParentID:                       nil,
		SharedRunnersMinutesLimit:      nil,
		ExtraSharedRunnersMinutesLimit: nil,
	})
	defer resp.Body.Close()

	if err := utilerrors.NewConflict(err); err == nil || errors.IsConflict(err) {
		if group == nil {
			if groups, err := g.list(workspace.Name); err != nil {
				g.logger.WithFields(workspaceLogInfo).WithFields(logrus.Fields{
					"message": "failed to get gitlab group",
				}).Error(err)
				return nil, err
			} else {
				group = groups[0]
				g.logger.WithFields(workspaceLogInfo).WithFields(logrus.Fields{
					"groupName": group.Name,
					"groupId":   group.ID,
				}).Info("group already exist, finish to get gitlab group")
			}
		}
		_, err := g.pagerClient.
			DevopsV1alpha1().
			Pagers(constants.DevopsNamespace).
			Create(g.ctx, &v1alpha1.Pager{
				ObjectMeta: v1.ObjectMeta{
					Name: "workspace-" + workspace.Name,
				},
				Spec: v1alpha1.PagerSpec{
					MessageID:   strconv.Itoa(group.ID),
					MessageName: group.Name,
					MessageType: workspace.Kind,
				},
			}, v1.CreateOptions{})
		if err == nil || errors.IsAlreadyExists(err) {
			return group, nil
		} else {
			g.logger.WithFields(workspaceLogInfo).WithFields(logrus.Fields{
				"message": "failed to create pager",
				"pager":   "workspace-" + workspace.Name,
			}).Error(err)
			return group, err
		}
	} else {
		g.logger.WithFields(workspaceLogInfo).WithFields(logrus.Fields{
			"message": "failed to create gitlab group",
		}).Error(err)
		return nil, err
	}
}

func (g groupInfo) Update(objOld interface{}, objNew interface{}) error {
	if objOld == nil {
		//this is an add operation
	}
	panic("implement me")
}

func (g groupInfo) Delete(workspaceName string) error {
	pagerName := "workspace-" + workspaceName

	workspaceLogInfo := logrus.Fields{
		"workspace": workspaceName,
	}
	g.logger.WithFields(workspaceLogInfo).Info("start to delete gitlab pager")

	err := g.pagerClient.DevopsV1alpha1().Pagers(constants.DevopsNamespace).Delete(g.ctx, pagerName, v1.DeleteOptions{})
	if err == nil || errors.IsNotFound(err) {
		g.logger.WithFields(workspaceLogInfo).WithFields(logrus.Fields{
			"pager": pagerName,
		}).Info("finish to delete gitlab group pager")
		return nil
	} else {
		g.logger.WithFields(workspaceLogInfo).WithFields(logrus.Fields{
			"message": "failed to delete gitlab group pager",
			"pager":   pagerName,
		}).Error(err)
		return err
	}
}

func (g groupInfo) GetByName(key string) (interface{}, error) {
	return g.list(key)
}

func (g groupInfo) GetByID(id int) (interface{}, error) {
	//g.gitClient.Groups.GetGroup()
	panic("implement me")
}

func (g groupInfo) List(key string) (interface{}, error) {
	groups, err := g.list(key)
	return groups, err
}

func (g groupInfo) list(key string) ([]*git.Group, error) {
	groups, resp, err := g.gitlabClient.Client.Groups.ListGroups(&git.ListGroupsOptions{
		Search: git.String(key),
	})
	defer resp.Body.Close()
	if err != nil {
		g.logger.WithFields(logrus.Fields{
			"event": "list",
			"msg":   resp.Body,
		}).Error(err)
		return nil, err
	} else {
		return groups, nil
	}
}

func NewGroupGenerator(name string, ctx context.Context, gitlabClient *clientset.GitlabClient, pagerClient *pager.Clientset) syncer.Generator {
	//cancelCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	logger := utils.GetLogger(logrus.Fields{
		"component": "gitlab",
		"resource":  "group",
	})
	return &groupInfo{
		groupName:    name,
		gitlabClient: gitlabClient,
		pagerClient:  pagerClient,
		ctx:          ctx,
		logger:       logger,
	}
}
