package service

import (
	"context"

	"github.com/keboola/temp-webhooks-api/internal/pkg/env"
	"github.com/keboola/temp-webhooks-api/internal/pkg/webhooks/api/gen/webhooks"
)

type Service struct {
	envs *env.Map
}

func New(envs *env.Map) webhooks.Service {
	return &Service{envs: envs}
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
