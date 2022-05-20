package gitlab

import (
	"context"
	baseErr "errors"
	"github.com/hchenc/iceberg/pkg/apis/iam/v1alpha2"
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

type userInfo struct {
	username     string
	password     string
	gitlabClient *clientset.GitlabClient
	pagerClient  *pager.Clientset
	logger       *logrus.Logger
	ctx          context.Context
}

func (u userInfo) Create(obj interface{}) (interface{}, error) {
	user := obj.(*v1alpha2.User)
	if exist, err := u.pagerClient.DevopsV1alpha1().Pagers(constants.DevopsNamespace).Get(u.ctx, "user-" + user.Name, v1.GetOptions{}); err == nil && exist != nil {
		return nil, nil
	}

	gitlabUser, resp, err := u.gitlabClient.Client.Users.CreateUser(&git.CreateUserOptions{
		Email:               git.String(user.Spec.Email),
		ResetPassword:       git.Bool(true),
		ForceRandomPassword: nil,
		Username:            git.String(user.Name),
		Name:                git.String(user.Name),
		Skype:               nil,
		Linkedin:            nil,
		Twitter:             nil,
		WebsiteURL:          nil,
		Organization:        nil,
		ProjectsLimit:       nil,
		ExternUID:           nil,
		Provider:            nil,
		Bio:                 nil,
		Location:            nil,
		Admin:               nil,
		CanCreateGroup:      git.Bool(false),
		SkipConfirmation:    nil,
		External:            nil,
		PrivateProfile:      nil,
		Note:                nil,
	})
	defer resp.Body.Close()
	if err := utilerrors.NewConflict(err); err == nil || errors.IsConflict(err) {
		if gitlabUser == nil {
			if gitlabUsers, err := u.list(user.Name); err != nil {
				return nil, err
			} else if len(gitlabUsers) != 0 {
				gitlabUser = gitlabUsers[0]
			} else {
				return nil, utilerrors.NewNotFound(baseErr.New("gitlab user not found"))
			}
		}
		_, err := u.pagerClient.
			DevopsV1alpha1().
			Pagers(constants.DevopsNamespace).
			Create(u.ctx, &v1alpha1.Pager{
				ObjectMeta: v1.ObjectMeta{
					Name: "user-" + user.Name,
				},
				Spec: v1alpha1.PagerSpec{
					MessageID:   strconv.Itoa(gitlabUser.ID),
					MessageName: gitlabUser.Name,
					MessageType: user.Kind,
				},
			}, v1.CreateOptions{})
		if err == nil || errors.IsAlreadyExists(err) {
			return gitlabUser, nil
		} else {
			return gitlabUser, err
		}
	} else {
		return nil, err
	}
}

func (u userInfo) Update(objOld interface{}, objNew interface{}) error {
	panic("implement me")
}

func (u userInfo) Delete(userName string) error {
	pagerName := "user-" + userName
	userLogInfo := logrus.Fields{
		"user": userName,
	}
	u.logger.WithFields(userLogInfo).Info("start to delete gitlab user pager")

	err := u.pagerClient.DevopsV1alpha1().Pagers(constants.DevopsNamespace).Delete(u.ctx, pagerName, v1.DeleteOptions{})
	if err == nil || errors.IsNotFound(err) {
		u.logger.WithFields(userLogInfo).WithFields(logrus.Fields{
			"pager": pagerName,
		}).Info("finish to delete gitlab user pager")
		return nil
	} else {
		u.logger.WithFields(userLogInfo).WithFields(logrus.Fields{
			"message": "failed to delete gitlab user pager",
			"pager":   pagerName,
		}).Error(err)
		return err
	}
}

func (u userInfo) GetByName(name string) (interface{}, error) {
	panic("implement me")
}

func (u userInfo) GetByID(id int) (interface{}, error) {
	panic("implement me")
}

func (u userInfo) List(key string) (interface{}, error) {
	panic("implement me")
}

func (u userInfo) list(key string) ([]*git.User, error) {
	users, resp, err := u.gitlabClient.Client.Users.ListUsers(&git.ListUsersOptions{
		Username: git.String(key),
	})
	defer resp.Body.Close()
	if err != nil {
		u.logger.WithFields(logrus.Fields{
			"event": "list",
		}).Error(err.Error())
		return nil, err
	} else {
		return users, nil
	}
}

func NewUserGenerator(ctx context.Context, client *clientset.GitlabClient, pageClient *pager.Clientset) syncer.Generator {
	logger := utils.GetLogger(logrus.Fields{
		"component": "gitlab",
		"resource":  "user",
	})
	return &userInfo{
		pagerClient:  pageClient,
		gitlabClient: client,
		ctx:          ctx,
		logger:       logger,
	}
}
