package clientset

import (
	"errors"
	"github.com/hchenc/iceberg/pkg/config"
	"github.com/hchenc/iceberg/pkg/constants"
	"github.com/xanzy/go-gitlab"
	"strings"
)

type GitlabClient struct {
	Client           *gitlab.Client
	Username         string
	Password         string
	IntegrateOptions []*config.IntegrateOption
}

func convertAppType(appType string) string {
	if len(appType) == 0 {
		appType = constants.DefaultPipeline
	} else {
		for _, value := range []string{
			"java",
			"python",
			"nodejs",
			"go",
		} {
			if strings.Contains(appType, value) {
				appType = value
				break
			}
		}
	}
	return appType
}

func (g *GitlabClient) GetIntegrateOption(appType string) (*config.IntegrateOption, error) {
	appType = convertAppType(appType)

	for _, integrateOption := range g.IntegrateOptions {
		if integrateOption.Pipeline == appType {
			return integrateOption, nil
		}
	}
	return nil, errors.New("pipeline not found")
}

func NewGitlabClient(gitlabOptions *config.GitlabOptions, integrateOptions []*config.IntegrateOption) *GitlabClient {
	var gitlabClient GitlabClient

	gc, err := gitlab.NewBasicAuthClient(gitlabOptions.User,
		gitlabOptions.Password,
		gitlab.WithBaseURL("http://"+gitlabOptions.Host+":"+gitlabOptions.Port),
		gitlab.WithoutRetries())
	if err != nil {
		//panic(err)
		return nil
	}

	gitlabClient.Client = gc
	gitlabClient.Username = gitlabOptions.User
	gitlabClient.Password = gitlabOptions.Password
	gitlabClient.IntegrateOptions = integrateOptions

	return &gitlabClient
}
