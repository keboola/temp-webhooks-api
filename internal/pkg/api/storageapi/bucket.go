package storageapi

import (
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/keboola/temp-webhooks-api/internal/pkg/http/client"
	"github.com/keboola/temp-webhooks-api/internal/pkg/model"
)

func (a *Api) CreateBucketAsync(name string, stage string, displayName string) (model.Job, error) {
	if stage == "" {
		stage = "in"
	}
	if stage != "in" && stage != "out" {
		return model.Job{}, fmt.Errorf("wrong stage, allowed values: in, out")
	}
	response := a.CreateBucketAsyncRequest(name, stage, displayName).Send().Response

	if response.HasResult() {
		return *response.Result().(*model.Job), nil
	}
	return model.Job{}, response.Err()
}

func (a *Api) CreateBucketAsyncRequest(name string, stage string, displayName string) *client.Request {
	job := &model.Job{}
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
		SetResult(job)
	request.
		OnSuccess(waitForJob(a, request, job, nil))
	return request
}

func (a *Api) CreateTableAsync(tableId string, tableName string, fileId string) (model.Job, error) {
	response := a.CreateTableAsyncRequest(tableId, tableName, fileId).Send().Response

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
