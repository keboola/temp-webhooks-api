package storage

import (
	"fmt"
	"sync"

	"github.com/keboola/temp-webhooks-api/internal/pkg/log"
	"github.com/keboola/temp-webhooks-api/internal/pkg/model"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"gorm.io/gorm"
)

type Storage struct {
	lock     *sync.Mutex
	db       *gorm.DB
	logger   log.Logger
	webhooks model.WebhooksMap
}

func New(db *gorm.DB, logger log.Logger) *Storage {
	return &Storage{
		lock:     &sync.Mutex{},
		db:       db,
		logger:   logger,
		webhooks: make(model.WebhooksMap),
	}
}

// ImportData imports data from memory to the table, if model.Conditions are meet.
func (s *Storage) ImportData() error {
	for _, webhook := range s.webhooks {
		// nolint: forbidigo
		fmt.Printf("please resolve conditions: %#v\n", webhook.Conditions)
	}
	return nil
}

func (s *Storage) Get(hash string) (*model.Webhook, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	webhook, found := s.webhooks[model.WebhookHash(hash)]
	if !found {
		return nil, fmt.Errorf(`webhook "%s" not found`, hash)
	}
	return webhook, nil
}

func (s *Storage) Register(token, tableId string, conditions model.Conditions) (*model.Webhook, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	hash := model.WebhookHash(gonanoid.Must())
	webhook, err := model.NewWebhook(token, tableId, hash, conditions)
	if err != nil {
		return nil, err
	}

	s.webhooks[hash] = webhook
	return webhook, nil
}
