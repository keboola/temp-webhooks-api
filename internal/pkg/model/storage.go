package model

import (
	"fmt"
	"sync"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

type Storage struct {
	lock     *sync.Mutex
	webhooks WebhooksMap
}

func NewStorage() *Storage {
	return &Storage{
		lock:     &sync.Mutex{},
		webhooks: make(WebhooksMap),
	}
}

func (s *Storage) Get(hash string) (*Webhook, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	webhook, found := s.webhooks[Hash(hash)]
	if !found {
		return nil, fmt.Errorf(`webhook "%s" not found`, hash)
	}
	return webhook, nil
}

func (s *Storage) Register(token, tableId string, conditions Conditions) (*Webhook, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	hash := Hash(gonanoid.Must())
	webhook, err := NewWebhook(token, tableId, hash, conditions)
	if err != nil {
		return nil, err
	}

	s.webhooks[hash] = webhook
	return webhook, nil
}
