package clientset

import (
	"context"
	harbor2 "github.com/hchenc/go-harbor"
	"github.com/hchenc/iceberg/pkg/config"
)

type HarborClient struct {
	*harbor2.APIClient
}

func NewHarborClient(harborOptions *config.HarborOptions, ctx context.Context) *HarborClient {
	basicAuth := harbor2.BasicAuth{
		UserName: harborOptions.User,
		Password: harborOptions.Password,
	}
	return &HarborClient{
		harbor2.NewAPIClient(
			harbor2.NewConfigurationWithContext(
				harborOptions.Host,
				context.WithValue(
					ctx,
					harbor2.ContextBasicAuth,
					basicAuth),
			),
		),
	}
}
