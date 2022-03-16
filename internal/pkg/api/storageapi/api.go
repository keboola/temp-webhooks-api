package storageapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/keboola/temp-webhooks-api/internal/pkg/http/client"
	"github.com/keboola/temp-webhooks-api/internal/pkg/log"
	"github.com/keboola/temp-webhooks-api/internal/pkg/model"
	"github.com/keboola/temp-webhooks-api/internal/pkg/utils"
)

type Api struct {
	apiHost    string
	apiHostUrl string
	client     *client.Client
	logger     log.Logger
	token      *model.Token
}

func NewWithToken(ctx context.Context, logger log.Logger, host, tokenStr string, verbose bool) (*Api, error) {
	if host == "" {
		panic(fmt.Errorf("api host is not set"))
	}
	if tokenStr == "" {
		panic(fmt.Errorf("api token is not set"))
	}

	storageApi := New(ctx, logger, host, verbose)
	token, err := storageApi.GetToken(tokenStr)
	if err != nil {
		var errWithResponse client.ErrorWithResponse
		if errors.As(err, &errWithResponse) && errWithResponse.IsUnauthorized() {
			return nil, fmt.Errorf("the specified storage API token is not valid")
		} else {
			return nil, utils.PrefixError("token verification failed", err)
		}
	}
	if !token.IsMaster {
		return nil, fmt.Errorf("required master token, but the given token is not master")
	}

	logger.Debugf("Storage API token is valid.")
	logger.Debugf(`Project id: "%d", project name: "%s".`, token.ProjectId(), token.ProjectName())
	return storageApi.WithToken(token), nil
}

func New(ctx context.Context, logger log.Logger, host string, verbose bool) *Api {
	if host == "" {
		panic(fmt.Errorf("api host is not set"))
	}
	apiHostUrl := "https://" + host + "/v2/storage"
	c := client.NewClient(ctx, logger, verbose).WithHostUrl(apiHostUrl)
	c.SetError(&Error{})
	api := &Api{client: c, logger: logger, apiHost: host, apiHostUrl: apiHostUrl}
	return api
}

func (a *Api) Host() string {
	if len(a.apiHost) == 0 {
		panic(fmt.Errorf("api host is not set"))
	}
	return a.apiHost
}

func (a *Api) HostUrl() string {
	if len(a.apiHost) == 0 {
		panic(fmt.Errorf("api host is not set"))
	}
	return a.apiHostUrl
}

func (a *Api) NewPool() *client.Pool {
	return a.client.NewPool(a.logger)
}

func (a *Api) NewRequest(method string, url string) *client.Request {
	return a.client.NewRequest(method, url)
}

func (a *Api) Send(request *client.Request) {
	a.client.Send(request)
}

func (a *Api) SetRetry(count int, waitTime time.Duration, maxWaitTime time.Duration) {
	a.client.SetRetry(count, waitTime, maxWaitTime)
}

func (a *Api) RestyClient() *resty.Client {
	return a.client.GetRestyClient()
}

func (a *Api) HttpClient() *http.Client {
	return a.client.GetRestyClient().GetClient()
}
