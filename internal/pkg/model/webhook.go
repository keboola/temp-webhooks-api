package model

import (
	"fmt"
	"sync"
)

type Hash string

type WebhooksMap map[Hash]*Webhook

type Webhook struct {
	lock       *sync.Mutex
	Token      string
	TableId    string
	Hash       Hash
	Conditions Conditions
	File       *CsvFile
}

func NewWebhook(token, tableId string, hash Hash, conditions Conditions) (*Webhook, error) {
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

func (v *Webhook) WaitingRows() int {
	return v.File.Rows()
}

func (v *Webhook) Url(host string) string {
	return fmt.Sprintf("https://%s/import/%s", host, v.Hash)
}
