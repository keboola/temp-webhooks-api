package storageapi

import (
	"github.com/go-resty/resty/v2"
	"github.com/keboola/temp-webhooks-api/internal/pkg/http/client"
	"github.com/keboola/temp-webhooks-api/internal/pkg/model"
)

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
