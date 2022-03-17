package storage

import (
	"errors"
	"fmt"
	"time"

	"github.com/keboola/temp-webhooks-api/internal/pkg/log"
	"github.com/keboola/temp-webhooks-api/internal/pkg/model"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"gorm.io/gorm"
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
	hash := model.WebhookHash(hashStr)
	webhook := model.Webhook{}
	err := s.db.First(&webhook, "hash = ?", hash).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf(`webhook "%s" not found`, hash)
	}
	return &webhook, err
}

func (s *Storage) Register(token, tableId string, conditions model.Conditions) (*model.Webhook, error) {
	hash := model.WebhookHash(gonanoid.Must())
	webhook := &model.Webhook{
		Hash:       hash,
		Token:      token,
		TableId:    tableId,
		Conditions: conditions,
	}
	return webhook, s.db.Create(webhook).Error
}

func (s *Storage) WriteRow(webhook *model.Webhook, headers, body string) error {
	row := &model.Row{
		Webhook: webhook.Id,
		Time:    time.Now(),
		Headers: headers,
		Body:    body,
	}
	if err := s.db.Create(row).Error; err != nil {
		return fmt.Errorf("cannot write data to db: %w", err)
	}
	return nil
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
