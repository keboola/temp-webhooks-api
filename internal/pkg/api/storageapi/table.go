package storageapi

import (
	"fmt"
	"net/http"

	"github.com/go-resty/resty/v2"
	"github.com/keboola/temp-webhooks-api/internal/pkg/http/client"
	"github.com/keboola/temp-webhooks-api/internal/pkg/model"
)

func (a *Api) TableExists(tableId string) bool {
	response := a.GetTableRequest(tableId).Send().Response
	return response.Response.StatusCode() == http.StatusOK
}

func (a *Api) GetTableRequest(tableId string) *client.Request {
	return a.NewRequest(resty.MethodGet, fmt.Sprintf("tables/%s", tableId))
}

func (a *Api) ImportTableAsync(tableId string, fileId string, incremental bool) (model.Job, error) {
	response := a.ImportTableAsyncRequest(tableId, fileId, incremental).Send().Response

	if response.HasResult() {
		return *response.Result().(*model.Job), nil
	}
	return model.Job{}, response.Err()
}

func (a *Api) ImportTableAsyncRequest(tableId string, fileId string, incremental bool) *client.Request {
	job := &model.Job{}
	body := map[string]string{
		"dataFileId": fileId,
	}
	if incremental {
		body["incremental"] = "1"
	}
	request := a.
		NewRequest(resty.MethodPost, fmt.Sprintf("tables/%s/import-async", tableId)).
		SetFormBody(body).
		SetResult(job)
	request.
		OnSuccess(waitForJob(a, request, job, nil))
	return request
}
