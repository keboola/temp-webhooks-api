package storage

import (
	"errors"
	"fmt"
	"time"

	"github.com/keboola/temp-webhooks-api/internal/pkg/log"
	"github.com/keboola/temp-webhooks-api/internal/pkg/model"
	"github.com/keboola/temp-webhooks-api/internal/pkg/webhooks/api/gen/webhooks"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Storage struct {
	db     *gorm.DB
	logger log.Logger
}

func New(db *gorm.DB, logger log.Logger) *Storage {
	return &Storage{
		db:     db,
		logger: logger,
	}
}

// ImportData imports data from memory to the table, if model.Conditions are meet.
func (s *Storage) ImportData() error {
	return nil
}

func (s *Storage) Get(hashStr string) (*model.Webhook, error) {
	return getWebhook(hashStr, s.db)
}

func (s *Storage) Register(token model.Token, tableId string, conditions model.Conditions) (*model.Webhook, error) {
	hash := model.WebhookHash(gonanoid.Must())
	webhook := &model.Webhook{
		Hash:       hash,
		ProjectId:  uint32(token.ProjectId()),
		Token:      token.Token,
		TableId:    tableId,
		Conditions: conditions,
	}
	return webhook, s.db.Create(webhook).Error
}

func (s *Storage) WriteRow(webhookHash string, headers, body string) (webhook *model.Webhook, count uint64, err error) {
	var countInt int64
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Get webhook, select for update
		webhook, err = getWebhook(webhookHash, tx.Clauses(clause.Locking{Strength: "UPDATE"}))
		if err != nil {
			return &webhooks.WebhookNotFoundError{Message: err.Error()}
		}

		// Create row
		row := &model.Row{
			Webhook: webhook.Id,
			Time:    time.Now(),
			Headers: headers,
			Body:    body,
		}

		// Insert row
		if err := tx.Create(row).Error; err != nil {
			return fmt.Errorf("cannot write data to db: %w", err)
		}

		// Update size
		size := uint64(len(headers) + len(body))
		tx.Model(&model.Webhook{}).Where("id = ?", webhook.Id).Update("size", webhook.Size+size)

		// Get current batch size
		if err := tx.Model(&model.Row{}).Where("webhook = ?", webhook.Id).Count(&countInt).Error; err != nil {
			return fmt.Errorf("cannot count rows: %w", err)
		}

		return nil
	})
	return webhook, uint64(countInt), err
}

func (s *Storage) MigrateDb() error {
	lockName := "__db_migration__"
	lockTimeout := 30
	if err := s.db.Exec(`SELECT GET_LOCK(?, ?)`, lockName, lockTimeout).Error; err != nil {
		return fmt.Errorf("db migration: cannot create lock: %w", err)
	}
	if err := s.db.AutoMigrate(&model.Webhook{}, &model.Row{}); err != nil {
		return fmt.Errorf("db migration: cannot migrate: %w", err)
	}
	if err := s.db.Exec(`SELECT RELEASE_LOCK(?)`, lockName).Error; err != nil {
		return fmt.Errorf("db migration: cannot release lock: %w", err)
	}
	return nil
}

func getWebhook(hashStr string, db *gorm.DB) (*model.Webhook, error) {
	hash := model.WebhookHash(hashStr)
	webhook := model.Webhook{}

	err := db.First(&webhook, "hash = ?", hash).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf(`webhook "%s" not found`, hash)
	}
	return &webhook, err
}
