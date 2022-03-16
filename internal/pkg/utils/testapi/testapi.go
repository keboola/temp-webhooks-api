package testapi

import (
	"context"
	"os"
	"time"

	"github.com/jarcoal/httpmock"
	"github.com/keboola/temp-webhooks-api/internal/pkg/api/storageapi"
	"github.com/keboola/temp-webhooks-api/internal/pkg/log"
	"github.com/keboola/temp-webhooks-api/internal/pkg/model"
)

func NewMockedStorageApi(logger log.DebugLogger) (*storageapi.Api, *httpmock.MockTransport) {
	// Set short retry delay in tests
	api := storageapi.New(context.Background(), logger, "connection.keboola.com", false)
	api.SetRetry(3, 1*time.Millisecond, 1*time.Millisecond)
	api = api.WithToken(model.Token{Owner: model.TokenOwner{Id: 12345}})

	// Mocked resty transport
	transport := httpmock.NewMockTransport()
	api.HttpClient().Transport = transport
	return api, transport
}

func NewStorageApi(host string, verbose bool) (*storageapi.Api, log.DebugLogger) {
	logger := log.NewDebugLogger()
	if verbose {
		logger.ConnectTo(os.Stdout)
	}
	a := storageapi.New(context.Background(), logger, host, false)
	a.SetRetry(3, 100*time.Millisecond, 100*time.Millisecond)
	return a, logger
}

func NewStorageApiWithToken(host, tokenStr string, verbose bool) (*storageapi.Api, log.DebugLogger) {
	a, logger := NewStorageApi(host, verbose)
	token, err := a.GetToken(tokenStr)
	if err != nil {
		panic(err)
	}
	return a.WithToken(token), logger
}
