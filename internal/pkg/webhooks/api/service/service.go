package service

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	stdLog "log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/keboola/temp-webhooks-api/internal/pkg/api/storageapi"
	"github.com/keboola/temp-webhooks-api/internal/pkg/env"
	"github.com/keboola/temp-webhooks-api/internal/pkg/json"
	"github.com/keboola/temp-webhooks-api/internal/pkg/log"
	"github.com/keboola/temp-webhooks-api/internal/pkg/model"
	"github.com/keboola/temp-webhooks-api/internal/pkg/s3"
	"github.com/keboola/temp-webhooks-api/internal/pkg/storage"
	"github.com/keboola/temp-webhooks-api/internal/pkg/webhooks/api/gen/webhooks"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

const (
	WebhookCheckInterval = 15 * time.Second
	HeadersCtxKey        = ctxKey("headers")
)

type ctxKey string

type Service struct {
	lock       *sync.Mutex
	updating   map[model.WebhookHash]bool
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

	// Create API
	api := storageapi.New(context.Background(), logger, storageApiHost, false)

	// Create service
	s := &Service{
		lock:       &sync.Mutex{},
		updating:   make(map[model.WebhookHash]bool),
		ctx:        ctx,
		host:       serviceHost,
		envs:       envs,
		logger:     logger,
		storage:    stg,
		storageApi: api,
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
	s.lock.Lock()
	defer s.lock.Unlock()

	// Get all
	items, err := s.storage.AllWebhooks()
	if err != nil {
		s.logger.Error(err)
	}

	// Check each
	for _, webhook := range items {
		// Only once
		if s.updating[webhook.Hash] {
			s.logger.Infof(`skipped import "%s": in progress`, webhook.Hash)
			continue
		}

		// Count rows
		count, err := s.storage.CountRows(webhook.Id)
		if err != nil {
			s.logger.Error(err.Error())
			continue
		}
		if count == 0 {
			s.logger.Infof(`skipped import "%s": count=0`, webhook.Hash)
			continue
		}

		// Check
		if webhook.Conditions.ShouldImport(count, time.Since(webhook.ImportedAt), webhook.Size) {
			s.updating[webhook.Hash] = true
			go func() {
				defer func() {
					s.updating[webhook.Hash] = false
				}()
				if err := s.importToKbc(string(webhook.Hash)); err != nil {
					s.logger.Errorf(`cannot import "%s": %w`, webhook.Hash, err)
					return
				}
				s.logger.Infof(`IMPORTED to "%s"`, webhook.Hash)
			}()
		} else {
			s.logger.Infof(`skipped import "%s": condition=false`, webhook.Hash)
			continue
		}
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

func (s *Service) Register(_ context.Context, payload *webhooks.RegisterPayload) (res *webhooks.RegistrationResult, err error) {
	// Validate token
	token, err := s.storageApi.GetToken(payload.Token)
	if err != nil {
		return nil, &webhooks.UnauthorizedError{Message: fmt.Sprintf(`Invalid storage token "%s" supplied.`, payload.Token)}
	}

	// Validate table ID
	parts := strings.Split(payload.TableID, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf(`invalid table ID: %s`, payload.TableID)
	}

	// Create conditions
	conditions, err := conditionsFromPayload(payload.Conditions)
	if err != nil {
		return nil, err
	}

	// Create webhook
	webhook, err := s.storage.RegisterWebhook(token, payload.TableID, conditions)
	if err != nil {
		return nil, err
	}

	// Return URL
	url := webhook.Url(s.host)
	s.logger.Infof("REGISTERED webhook, tableId=\"%s\"", webhook.TableId)
	return &webhooks.RegistrationResult{URL: url}, nil
}

func (s *Service) Update(_ context.Context, payload *webhooks.UpdatePayload) (res *webhooks.UpdateResult, err error) {
	// Create conditions
	conditions, err := conditionsFromPayload(payload.Conditions)
	if err != nil {
		return nil, err
	}

	webhook, err := s.storage.UpdateWebhook(payload.Hash, conditions)
	if err != nil {
		return nil, err
	}
	return &webhooks.UpdateResult{Conditions: webhook.Conditions.Payload()}, nil
}

func (s *Service) Flush(_ context.Context, payload *webhooks.FlushPayload) (res string, err error) {
	// Import to KBC
	if err = s.importToKbc(payload.Hash); err != nil {
		return "", err
	}
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
	webhook, count, err := s.storage.WriteRow(payload.Hash, headers, string(body))
	if err != nil {
		return nil, err
	}

	s.logger.Infof("RECEIVED webhook, tableId=\"%s\"", webhook.TableId)
	return &webhooks.ImportResult{RecordsInBatch: count}, nil
}

func (s *Service) importToKbc(webhookHash string) error {
	// Get webhook
	webhook, err := s.storage.Get(webhookHash)
	if err != nil {
		return err
	}

	// Parse tableID
	parts := strings.Split(webhook.TableId, ".")
	if len(parts) != 3 {
		return fmt.Errorf(`invalid table ID: %s`, webhook.TableId)
	}
	bucketId := strings.Join(parts[0:2], ".")
	tableName := parts[2]

	// Set token
	apiWithToken := s.storageApi.WithToken(model.Token{Token: webhook.Token})

	// Create bucket if not exists
	if !apiWithToken.BucketExists(bucketId) {
		bucketName := strings.TrimPrefix(parts[1], "c-")
		if _, err := apiWithToken.CreateBucket(bucketName, parts[0], parts[1]); err != nil {
			return fmt.Errorf(`cannot create bucket "%s": %w`, bucketId, err)
		}
		s.logger.Infof(`created bucket "%s"`, bucketId)
	} else {
		s.logger.Infof(`bucket "%s" exists`, bucketId)
	}

	// Create temp file
	csvFile, err := ioutil.TempFile(os.TempDir(), "keboola-csv")
	if err != nil {
		return fmt.Errorf(`cannot create temp file: %w`, err)
	}
	defer func() {
		if err := os.Remove(csvFile.Name()); err != nil {
			s.logger.Error(err.Error())
		}
	}()

	// Create CSV file
	if _, err = s.storage.Fetch(webhookHash, csvFile); err != nil {
		return err
	}
	s.logger.Infof(`fetched "%s" to CSV file "%s"`, webhook.Hash, csvFile.Name())

	// Create file resource
	fileResource, err := apiWithToken.CreateFileResource(fmt.Sprintf("webhook-%s.csv", webhook.Hash))
	if err != nil {
		return fmt.Errorf(`cannot create file resource: %w`, err)
	}

	// Upload to S3
	err = s3.UploadFileToS3(csvFile.Name(), fileResource)
	if err != nil {
		return fmt.Errorf(`cannot upload to S3: %w`, err)
	}

	// Import CSV
	fileId := strconv.Itoa(fileResource.Id)
	if apiWithToken.TableExists(webhook.TableId) {
		// Import table
		_, err = apiWithToken.ImportTableAsync(webhook.TableId, fileId, true)
		if err != nil {
			return fmt.Errorf(`cannot import to table "%s": %w`, webhook.TableId, err)
		}
	} else {
		// Create table
		_, err = apiWithToken.CreateTableAsync(bucketId, tableName, fileId)
		if err != nil {
			return fmt.Errorf(`cannot create table "%s": %w`, webhook.TableId, err)
		}
	}

	return nil
}

func conditionsFromPayload(payload *webhooks.Conditions) (model.Conditions, error) {
	// Create conditions
	conditions := model.NewConditions()
	if payload != nil {
		if err := conditions.SetCount(payload.Count); err != nil {
			return conditions, err
		}
		if err := conditions.SetTime(payload.Time); err != nil {
			return conditions, err
		}
		if err := conditions.SetSize(payload.Size); err != nil {
			return conditions, err
		}
	}
	return conditions, nil
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
