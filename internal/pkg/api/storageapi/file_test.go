package storageapi_test

import (
	"testing"

	"github.com/keboola/temp-webhooks-api/internal/pkg/env"
	"github.com/keboola/temp-webhooks-api/internal/pkg/utils/testproject"
)

func TestX(t *testing.T) {
	t.Parallel()
	project := testproject.GetTestProject(t, env.Empty())
	_ = project.StorageApi()
}
