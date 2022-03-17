package model

import (
	"fmt"
	"time"
)

type WebhookHash string

type Webhook struct {
	Id         uint32      `gorm:"primaryKey;autoIncrement"`
	Hash       WebhookHash `gorm:"type:CHAR(21);index;not null"`
	ProjectId  uint32
	Token      string `gorm:"type:VARCHAR(255);not null"`
	TableId    string `gorm:"type:VARCHAR(1000);not null"`
	Size       uint64
	Conditions Conditions `gorm:"embedded;embeddedPrefix:condition_"`
	Data       []Row      `gorm:"foreignKey:Webhook"` // only for FK definition
}

func (v *Webhook) Url(host string) string {
	return fmt.Sprintf("https://%s/import/%s", host, v.Hash)
}

type Row struct {
	Webhook uint32
	Time    time.Time `gorm:"not null"`
	Headers string    `gorm:"not null"`
	Body    string    `gorm:"not null"`
}

func (Row) TableName() string {
	return "data"
}
