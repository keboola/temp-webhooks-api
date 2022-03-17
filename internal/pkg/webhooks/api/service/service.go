package service

import (
	"context"
	"fmt"
	"io"
	stdLog "log"
	"net/http"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/keboola/temp-webhooks-api/internal/pkg/api/storageapi"
	"github.com/keboola/temp-webhooks-api/internal/pkg/env"
	"github.com/keboola/temp-webhooks-api/internal/pkg/json"
	"github.com/keboola/temp-webhooks-api/internal/pkg/log"
	"github.com/keboola/temp-webhooks-api/internal/pkg/model"
	"github.com/keboola/temp-webhooks-api/internal/pkg/storage"
	"github.com/keboola/temp-webhooks-api/internal/pkg/webhooks/api/gen/webhooks"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

const (
	WebhookCheckInterval = 5 * time.Second
	HeadersCtxKey        = ctxKey("headers")
)

type ctxKey string

type Service struct {
	ctx        context.Context
	host       string
	envs       *env.Map
	logger     log.Logger
	storage    *storage.Storage
	storageApi *storageapi.Api
}

func New(ctx context.Context, envs *env.Map, stdLogger *stdLog.Logger) (webhooks.Service, error) {
	logger := log.NewApiLogger(stdLogger, "", false)

	// Load ENVs
	storageApiHost := envs.MustGet("KBC_STORAGE_API_HOST")
	serviceHost := envs.MustGet("SERVICE_HOST")
	mysqlDsn := envs.MustGet("SERVICE_MYSQL_DSN")

	// Connect to DB
	db, err := connectToDb(mysqlDsn, stdLogger)
	if err != nil {
		return nil, err
	}

	// Migrate DB
	stg := storage.New(db, logger)
	if err := stg.MigrateDb(); err != nil {
		return nil, err
	}

	// Create service
	s := &Service{
		ctx:        ctx,
		host:       serviceHost,
		envs:       envs,
		logger:     logger,
		storage:    stg,
		storageApi: storageapi.New(context.Background(), logger, storageApiHost, false),
	}
	s.StartCron()
	return s, nil
}

func (s *Service) StartCron() {
	go func() {
		ticker := time.NewTicker(WebhookCheckInterval)
		for {
			select {
			case <-s.ctx.Done():
				return
			case <-ticker.C:
				s.checkWebhooks()
			}
		}
	}()
}

// checkWebhooks checks if webhook should be imported. See model.Conditions.
func (s *Service) checkWebhooks() {
	if err := s.storage.ImportData(); err != nil {
		s.logger.Error(err)
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

func (s *Service) Import(ctx context.Context, payload *webhooks.ImportPayload, bodyStream io.ReadCloser) (res *webhooks.ImportResult, err error) {
	// Read body
	body, err := io.ReadAll(bodyStream)
	if err != nil {
		return nil, fmt.Errorf("cannot read request body: %w", err)
	}

	// Write CSV row
	headers := json.MustEncodeString(ctx.Value(HeadersCtxKey).(http.Header), true)
	webhook, err := s.storage.WriteRow(payload.Hash, headers, string(body))
	if err != nil {
		return nil, err
	}

	s.logger.Infof("RECEIVED webhook, tableId=\"%s\"", webhook.TableId)
	return &webhooks.ImportResult{RecordsInBatch: 0}, nil
}

func (s *Service) Register(_ context.Context, payload *webhooks.RegisterPayload) (res *webhooks.RegistrationResult, err error) {
	// Validate token
	token, err := s.storageApi.GetToken(payload.Token)
	if err != nil {
		return nil, err
	}

	// Create conditions
	conditions := model.NewConditions()
	if payload.Conditions != nil {
		// conditions.SetCount(payload.Conditions.Count)
		// conditions.SetTime(payload.Conditions.Time)
		// conditions.SetSize(payload.Conditions.Size)
	}

	// Create webhook
	webhook, err := s.storage.Register(token, payload.TableID, conditions)
	if err != nil {
		return nil, err
	}

	// Return URL
	url := webhook.Url(s.host)
	s.logger.Infof("REGISTERED webhook, tableId=\"%s\"", webhook.TableId)
	return &webhooks.RegistrationResult{URL: url}, nil
}

func connectToDb(mysqlDsn string, logger *stdLog.Logger) (db *gorm.DB, err error) {
	// Prepare
	dsn := mysqlDsn + "?timeout=10s&charset=utf8mb4&parseTime=True&loc=UTC"
	dbLogger := gormLogger.New(logger, gormLogger.Config{Colorful: false})
	dbConfig := &gorm.Config{Logger: dbLogger}

	// Connect with retry
	err = retry.Do(func() error {
		db, err = gorm.Open(mysql.Open(dsn), dbConfig)
		return err
	}, retry.Attempts(10), retry.Delay(2*time.Second), retry.DelayType(retry.FixedDelay))

	// Log
	if err == nil {
		logger.Printf(`DB connected to database "%s"`, db.Name())
	}
	return
}
