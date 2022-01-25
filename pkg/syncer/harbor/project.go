package harbor

import (
	harbor2 "github.com/hchenc/go-harbor"
	"github.com/hchenc/iceberg/pkg/apis/tenant/v1alpha2"
	"github.com/hchenc/iceberg/pkg/clients/clientset"
	"github.com/hchenc/iceberg/pkg/syncer"
	"github.com/hchenc/iceberg/pkg/utils"
	utilerrors "github.com/hchenc/iceberg/pkg/utils/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	"strconv"
)

type projectInfo struct {
	harborClient *clientset.HarborClient
	logger       *logrus.Logger
	username     string
	password     string
	host         string
}

func (p projectInfo) Create(obj interface{}) (interface{}, error) {
	workspace := obj.(*v1alpha2.WorkspaceTemplate)
	resp, err := p.harborClient.ProjectApi.CreateProject(harbor2.ProjectReq{
		ProjectName: workspace.Name,
		Metadata: &harbor2.ProjectMetadata{
			Public: "true",
		},
		StorageLimit: 0,
	}, &harbor2.ProjectApiCreateProjectOpts{})
	//defer resp.Body.Close()
	if err := utilerrors.NewConflict(err); err == nil || errors.IsConflict(err) {
		return resp, nil
	} else {
		return nil, nil
	}
}

func (p projectInfo) Update(objOld interface{}, objNew interface{}) error {
	panic("implement me")
}

func (p projectInfo) Delete(name string) error {
	panic("implement me")
}

func (p projectInfo) GetByName(name string) (interface{}, error) {
	project, resp, err := p.harborClient.ProjectApi.GetProject(name, &harbor2.ProjectApiGetProjectOpts{})
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	} else {
		return project, nil
	}
}

func (p projectInfo) GetByID(id int) (interface{}, error) {
	project, resp, err := p.harborClient.ProjectApi.GetProject(strconv.Itoa(id), &harbor2.ProjectApiGetProjectOpts{})
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	} else {
		return project, nil
	}
}

func (p projectInfo) List(key string) (interface{}, error) {
	panic("implement me")
}

func NewHarborProjectGenerator(name, group string, harborClient *clientset.HarborClient) syncer.Generator {
	logger := utils.GetLogger(logrus.Fields{
		"component": "harbor",
		"resource":  "project",
	})
	return &projectInfo{
		harborClient: harborClient,
		logger:       logger,
	}
}
