package storage

import (
	"errors"
	"fmt"

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

func (s *Storage) Get(hash string) (*model.Webhook, error) {
	webhook := model.Webhook{}
	err := s.db.First(&webhook, model.WebhookHash(hash)).Error
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
