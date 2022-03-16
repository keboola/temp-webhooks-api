package model

import (
	"fmt"
	"sync"
)

type WebhookHash string

type WebhooksMap map[WebhookHash]*Webhook

type Webhook struct {
	lock       *sync.Mutex
	Token      string
	TableId    string
	Hash       WebhookHash
	Conditions Conditions
	File       *CsvFile
}

func NewWebhook(token, tableId string, hash WebhookHash, conditions Conditions) (*Webhook, error) {
	f, err := NewCsvFile([]string{"Body"})
	if err != nil {
		return nil, err
	}
	return &Webhook{
		lock:       &sync.Mutex{},
		Token:      token,
		TableId:    tableId,
		Hash:       hash,
		Conditions: conditions,
		File:       f,
	}, nil
}

func (v *Webhook) WriteRow(record []string) error {
	return v.File.Write(record)
}

func (v *Webhook) WaitingRecords() int {
	return v.File.Rows()
}

func (v *Webhook) Url(host string) string {
	return fmt.Sprintf("https://%s/import/%s", host, v.Hash)
}
