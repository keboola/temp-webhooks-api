package model

import (
	"fmt"

	gonanoid "github.com/matoous/go-nanoid/v2"
)

type Storage struct {
	webhooks WebhooksMap
}

type Hash string

type WebhooksMap map[Hash]*Webhook

type Conditions struct {
	Count int
	Time  string
	Size  string
}

type Webhook struct {
	Token      string
	TableId    string
	Hash       Hash
	Conditions Conditions
}

func NewConditions() Conditions {
	return Conditions{
		Count: 1000,
		Time:  "30s",
		Size:  "10MB",
	}
}

func NewStorage() *Storage {
	return &Storage{
		webhooks: make(WebhooksMap),
	}
}

func (s *Storage) RegisterWebhook(token, tableId string, conditions Conditions) *Webhook {
	hash := Hash(gonanoid.Must())
	webhook := &Webhook{
		Token:      token,
		TableId:    tableId,
		Hash:       hash,
		Conditions: conditions,
	}
	s.webhooks[hash] = webhook
	return webhook
}

func (v *Webhook) Url(host string) string {
	return fmt.Sprintf("https://%s/import/%s", host, v.Hash)
}
