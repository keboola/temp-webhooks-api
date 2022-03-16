package storageapi_test

import (
	"testing"

	"github.com/keboola/temp-webhooks-api/internal/pkg/env"
	"github.com/keboola/temp-webhooks-api/internal/pkg/utils/testproject"
	"github.com/stretchr/testify/assert"
)

func TestPostCreateFileResource(t *testing.T) {
	t.Parallel()
	project := testproject.GetTestProject(t, env.Empty())
	api := project.StorageApi()
	response, err := api.CreateFileResource("tmpfile")
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Fail(t, "abc")
}
