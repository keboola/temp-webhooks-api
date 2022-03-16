package service

import (
	"context"
	stdLog "log"

	"github.com/keboola/temp-webhooks-api/internal/pkg/api/storageapi"
	"github.com/keboola/temp-webhooks-api/internal/pkg/env"
	"github.com/keboola/temp-webhooks-api/internal/pkg/log"
	"github.com/keboola/temp-webhooks-api/internal/pkg/model"
	"github.com/keboola/temp-webhooks-api/internal/pkg/webhooks/api/gen/webhooks"
)

type Service struct {
	host       string
	envs       *env.Map
	logger     log.Logger
	storage    *model.Storage
	storageApi *storageapi.Api
}

func New(envs *env.Map, stdLogger *stdLog.Logger) webhooks.Service {
	logger := log.NewApiLogger(stdLogger, "", false)
	storageApiHost := envs.MustGet("KBC_STORAGE_API_HOST")
	return &Service{
		host:       "localhost:8888",
		envs:       envs,
		logger:     logger,
		storage:    model.NewStorage(),
		storageApi: storageapi.New(context.Background(), logger, storageApiHost, false),
	}
}

func (s *Service) IndexRoot(_ context.Context) (res *webhooks.Index, err error) {
	res = &webhooks.Index{
		API:           "webhooks",
		Documentation: "https://webhooks.keboola.com/documentation",
	}
	return res, nil
}

func (s *Service) HealthCheck(_ context.Context) (res string, err error) {
	return "OK", nil
}

func (s *Service) Import(_ context.Context, payload *webhooks.ImportPayload) (res *webhooks.ImportResult, err error) {
	// Get webhook definition
	webhook, err := s.storage.Get(payload.Hash)
	if err != nil {
		return nil, err
	}

	// Write CSV row
	if err := webhook.WriteRow([]string{payload.Body}); err != nil {
		return nil, err
	}

	s.logger.Infof("RECEIVED webhook, tableId=\"%s\"", webhook.TableId)
	return &webhooks.ImportResult{WaitingForImport: webhook.WaitingRows()}, nil
}

func (s *Service) Register(_ context.Context, payload *webhooks.RegisterPayload) (res *webhooks.RegistrationResult, err error) {
	// Validate token
	if _, err := s.storageApi.GetToken(payload.Token); err != nil {
		return nil, err
	}

	// Create conditions
	conditions := model.NewConditions()
	if payload.Conditions != nil {
		conditions.Count = payload.Conditions.Count
		conditions.Time = payload.Conditions.Time
		conditions.Size = payload.Conditions.Size
	}

	// Create webhook
	webhook, err := s.storage.Register(payload.Token, payload.TableID, conditions)
	if err != nil {
		return nil, err
	}

	// Return URL
	url := webhook.Url(s.host)
	s.logger.Infof("REGISTERED webhook, tableId=\"%s\"", webhook.TableId)
	return &webhooks.RegistrationResult{URL: url}, nil
}
