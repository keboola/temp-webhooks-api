package model

import (
	"fmt"
)

type WebhookHash string

type Webhook struct {
	Hash       WebhookHash `gorm:"primaryKey"`
	Conditions Conditions  `gorm:"foreignKey:Webhook"`
	Token      string
	TableId    string
}

func (v *Webhook) Url(host string) string {
	return fmt.Sprintf("https://%s/import/%s", host, v.Hash)
}
