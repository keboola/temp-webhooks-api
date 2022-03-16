package storageapi

import (
	"github.com/go-resty/resty/v2"
	"github.com/keboola/temp-webhooks-api/internal/pkg/http/client"
	"github.com/keboola/temp-webhooks-api/internal/pkg/model"
)

func (a *Api) PostTableImportAsync(tableId string, fileId string) (model.Job, error) {
	response := a.PostTableImportAsyncRequest(tableId, fileId).Send().Response

	if response.HasResult() {
		return *response.Result().(*model.Job), nil
	}
	return model.Job{}, response.Err()
}

func (a *Api) PostTableImportAsyncRequest(tableId string, fileId string) *client.Request {
	job := &model.Job{}
	request := a.
		NewRequest(resty.MethodPost, fmt.Sprintf("tables/%s/import-async", tableId)).
		SetFormBody(map[string]string{
			"dataFileId": fileId,
		}).
		SetResult(job)
	request.
		OnSuccess(waitForJob(a, request, job, nil))
	return request
}

func (a *Api) CreateFileResource(name string) (model.FileResource, error) {
	response := a.PostCreateFileResource(name).Send().Response

	if response.HasResult() {
		return *response.Result().(*model.FileResource), nil
	}
	return model.FileResource{}, response.Err()
}

func (a *Api) PostCreateFileResource(name string) *client.Request {
	return a.
		NewRequest(resty.MethodPost, "files/prepare").
		SetFormBody(map[string]string{
			"name":            name,
			"federationToken": "true",
		}).
		SetResult(&model.FileResource{})
}
