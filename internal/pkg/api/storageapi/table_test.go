package storageapi_test

import (
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/keboola/temp-webhooks-api/internal/pkg/env"
	"github.com/keboola/temp-webhooks-api/internal/pkg/s3"
	"github.com/keboola/temp-webhooks-api/internal/pkg/utils/testproject"
	"github.com/stretchr/testify/assert"
)

func TestPostCreateTable(t *testing.T) {
	t.Parallel()
	project := testproject.GetTestProject(t, env.Empty())
	api := project.StorageApi()

	response, err := api.CreateFileResource("tmpfile")
	assert.NoError(t, err)

	err = os.WriteFile("/tmp/dat1.csv", []byte("col1,col2,col3\ntest1,test2,test3\n"), 0o666)
	assert.NoError(t, err)
	assert.NotNil(t, response.Id)
	fileId := response.Id

	err = s3.UploadFileToS3("/tmp/dat1.csv", response)
	assert.NoError(t, err)

	bucketName := fmt.Sprintf("test%d", int(time.Now().UnixNano()))
	tableName := fmt.Sprintf("table-%d", int(time.Now().UnixNano()))

	assert.False(t, api.BucketExists(fmt.Sprintf("in.c-%s", bucketName)))

	bucket, err := api.CreateBucket(bucketName, "in", "")
	assert.NoError(t, err)

	assert.True(t, api.BucketExists(bucket.Id))

	tableId := fmt.Sprintf("%s.%s", bucket.Id, tableName)
	_, err = api.CreateTableAsync(tableId, tableName, strconv.Itoa(fileId))
	assert.NoError(t, err)

	_, err = api.ImportTableAsync(tableId, strconv.Itoa(fileId), false)
	assert.NoError(t, err)
}
