package s3

import (
	"os"
	"testing"

	"github.com/keboola/temp-webhooks-api/internal/pkg/env"
	"github.com/keboola/temp-webhooks-api/internal/pkg/utils/testproject"
	"github.com/stretchr/testify/assert"
)

func TestUpload(t *testing.T) {
	t.Parallel()
	project := testproject.GetTestProject(t, env.Empty())
	api := project.StorageApi()
	response, err := api.CreateFileResource("tmpfile")
	assert.NoError(t, err)

	err = os.WriteFile("/tmp/dat1", []byte("hello\ngo\n"), 0o666)
	assert.NoError(t, err)

	err = UploadFileToS3("/tmp/dat1", response)

	assert.NoError(t, err)
}
