package harbor

import (
	"context"
	"fmt"
	harbor2 "github.com/hchenc/go-harbor"
	"github.com/hchenc/iceberg/pkg/apis/tenant/v1alpha2"
	typesv1alpha1 "github.com/hchenc/iceberg/pkg/apis/types/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestCreateProject(t *testing.T) {
	workspaceTemplate := &v1alpha2.WorkspaceTemplate{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name: "test2",
		},
		Spec: typesv1alpha1.FederatedWorkspaceSpec{},
	}
	ctx := context.WithValue(context.Background(), harbor2.ContextBasicAuth, harbor2.BasicAuth{
		UserName: "admin",
		Password: "Harbor12345",
	})
	config := harbor2.NewConfigurationWithContext("http://harbor.hchenc.com:5088/api/v2.0", ctx)
	client := harbor2.NewAPIClient(config)

	harborGenerator := NewHarborProjectGenerator("", "", client)
	result, err := harborGenerator.Create(workspaceTemplate)
	fmt.Println(result, err)
}
