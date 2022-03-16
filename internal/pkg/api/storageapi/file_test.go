package storageapi

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/keboola/temp-webhooks-api/internal/pkg/env"
	"github.com/keboola/temp-webhooks-api/internal/pkg/utils/testproject"
)

func TestX(t *testing.T) {
	t.Parallel()
	project := testproject.GetTestProject(t, env.Empty())
	api := project.StorageApi()
	var reponse = api.PostCreateFileResource("tmpfile")
	assert.NotNil(t, reponse)
}
