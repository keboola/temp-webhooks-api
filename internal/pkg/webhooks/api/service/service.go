package service

import (
	"context"
	"log"

	"github.com/keboola/temp-webhooks-api/internal/pkg/env"
	"github.com/keboola/temp-webhooks-api/internal/pkg/model"
	"github.com/keboola/temp-webhooks-api/internal/pkg/webhooks/api/gen/webhooks"
)

type Service struct {
	host    string
	envs    *env.Map
	logger  *log.Logger
	storage *model.Storage
}

func New(envs *env.Map, logger *log.Logger) webhooks.Service {
	return &Service{
		host:    "localhost:8888",
		envs:    envs,
		logger:  logger,
		storage: model.NewStorage(),
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

func (s *Service) Import(ctx context.Context, payload *webhooks.ImportPayload) (res string, err error) {
	return "OK", nil
}

func (s *Service) Register(_ context.Context, payload *webhooks.RegisterPayload) (res *webhooks.Registration, err error) {
	conditions := model.NewConditions()
	if payload.Conditions != nil {
		conditions.Count = payload.Conditions.Count
		conditions.Time = payload.Conditions.Time
		conditions.Size = payload.Conditions.Size
	}

	webhook := s.storage.RegisterWebhook(payload.Token, payload.TableID, conditions)
	url := webhook.Url(s.host)
	s.logger.Printf("REGISTERED tableId=\"%s\", url=\"%s\"", webhook.TableId, url)
	return &webhooks.Registration{URL: url}, nil
}
