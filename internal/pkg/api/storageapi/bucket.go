package storageapi

import (
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
	"github.com/keboola/temp-webhooks-api/internal/pkg/http/client"
	"github.com/keboola/temp-webhooks-api/internal/pkg/model"
)

func (a *Api) BucketExists(bucketId string) bool {
	response := a.GetBucketRequest(bucketId).Send().Response
	return response.Response.StatusCode() == http.StatusOK
}

func (a *Api) GetBucketRequest(bucketId string) *client.Request {
	return a.NewRequest(resty.MethodGet, fmt.Sprintf("buckets/%s", bucketId))
}

func (a *Api) CreateBucket(name string, stage string, displayName string) (model.Bucket, error) {
	if stage == "" {
		stage = "in"
	}
	if stage != "in" && stage != "out" {
		return model.Bucket{}, fmt.Errorf("wrong stage, allowed values: in, out")
	}
	response := a.CreateBucketAsyncRequest(name, stage, displayName).Send().Response

	if response.HasResult() {
		return *response.Result().(*model.Bucket), nil
	}
	return model.Bucket{}, response.Err()
}

func (a *Api) CreateBucketAsyncRequest(name string, stage string, displayName string) *client.Request {
	bucket := &model.Bucket{}
	body := map[string]string{
		"name":  name,
		"stage": stage,
	}
	if displayName != "" {
		body["displayName"] = displayName
	}
	request := a.
		NewRequest(resty.MethodPost, "buckets").
		SetFormBody(body).
		SetResult(bucket)
	return request
}

func (a *Api) CreateTableAsync(bucketId string, tableName string, fileId string) (model.Job, error) {
	response := a.CreateTableAsyncRequest(bucketId, tableName, fileId).Send().Response

	if response.HasResult() {
		return *response.Result().(*model.Job), nil
	}
	return model.Job{}, response.Err()
}

func (a *Api) CreateTableAsyncRequest(bucketId string, tableName string, fileId string) *client.Request {
	job := &model.Job{}
	request := a.
		NewRequest(resty.MethodPost, fmt.Sprintf("buckets/%s/tables-async", bucketId)).
		SetFormBody(map[string]string{
			"name":       tableName,
			"dataFileId": fileId,
		}).
		SetResult(job)
	request.
		OnSuccess(waitForJob(a, request, job, nil))
	return request
}
