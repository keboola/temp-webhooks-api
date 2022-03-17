package storage

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
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

func (s *Storage) AllWebhooks() (webhooks []*model.Webhook, err error) {
	return webhooks, s.db.Find(webhooks).Error
}

func (s *Storage) Get(hashStr string) (*model.Webhook, error) {
	return getWebhook(hashStr, s.db)
}

func (s *Storage) CountRows(webhookId uint32) (uint, error) {
	return countRows(webhookId, s.db)
}

func (s *Storage) RegisterWebhook(token model.Token, tableId string, conditions model.Conditions) (*model.Webhook, error) {
	hash := model.WebhookHash(gonanoid.Must())
	webhook := &model.Webhook{
		Hash:       hash,
		ProjectId:  uint32(token.ProjectId()),
		Token:      token.Token,
		TableId:    tableId,
		ImportedAt: time.Now(),
		Size:       0,
		Conditions: conditions,
	}
	return webhook, s.db.Create(webhook).Error
}

func (s *Storage) UpdateWebhook(webhookHash string, conditions model.Conditions) (webhook *model.Webhook, err error) {
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Get webhook, select for update
		webhook, err = getWebhook(webhookHash, tx.Clauses(clause.Locking{Strength: "UPDATE"}))
		if err != nil {
			return err
		}

		// Update
		if err := tx.Model(&model.Webhook{}).Where("id = ?", webhook.Id).Updates(&model.Webhook{Conditions: conditions}).Error; err != nil {
			return err
		}

		// Load new values
		webhook, err = getWebhook(webhookHash, tx.Clauses(clause.Locking{Strength: "UPDATE"}))
		return err
	})
	return webhook, err
}

func (s *Storage) FlushData(webhookHash string) (result string, err error) {
	// Get webhook, select for update
	err = s.db.Transaction(func(tx *gorm.DB) error {
		_, err := getWebhook(webhookHash, tx.Clauses(clause.Locking{Strength: "SELECT"}))
		if err != nil {
			return err
		}
		return err
	})
	return "ok", err
}

func (s *Storage) WriteRow(webhookHash string, headers, body string) (webhook *model.Webhook, count uint, err error) {
	var countInt int64
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Get webhook, select for update
		webhook, err = getWebhook(webhookHash, tx.Clauses(clause.Locking{Strength: "UPDATE"}))
		if err != nil {
			return err
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
		if err := tx.Model(&model.Webhook{}).Where("id = ?", webhook.Id).Update("size", webhook.Size+size).Error; err != nil {
			return err
		}

		// Get current batch size
		count, err = countRows(webhook.Id, tx)
		if err := tx.Model(&model.Row{}).Where("webhook = ?", webhook.Id).Count(&countInt).Error; err != nil {
			return fmt.Errorf("cannot count rows: %w", err)
		}

		return nil
	})
	return webhook, uint(countInt), err
}

func (s *Storage) Fetch(webhookHash string, target io.Writer) (webhook *model.Webhook, err error) {
	csvWriter := csv.NewWriter(target)
	defer csvWriter.Flush()

	// Write header
	if err := csvWriter.Write([]string{"timestamp", "headers", "body"}); err != nil {
		return nil, err
	}

	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Get webhook, select for update
		webhook, err = getWebhook(webhookHash, tx.Clauses(clause.Locking{Strength: "UPDATE"}))
		if err != nil {
			return err
		}

		// Select rows
		rows, err := tx.Table("data").Where("webhook = ?", webhook.Id).Order("time").Rows()
		if err != nil {
			return err
		}
		defer rows.Close()

		// Load rows
		for rows.Next() {
			row := &model.Row{}
			if err := tx.ScanRows(rows, row); err != nil {
				return err
			}

			csvRow := []string{row.Time.Format(time.RFC3339), row.Headers, row.Body}
			if err := csvWriter.Write(csvRow); err != nil {
				return err
			}
		}

		if err := rows.Err(); err != nil {
			return err
		}

		// Clear rows
		if err := tx.Table("data").Where("webhook = ?", webhook.Id).Delete(&model.Row{}).Error; err != nil {
			return err
		}

		// Update size and importedAt
		if err := tx.Model(&model.Webhook{}).Where("id = ?", webhook.Id).Updates(&model.Webhook{Size: 0, ImportedAt: time.Now()}).Error; err != nil {
			return err
		}

		return nil
	})
	return webhook, err
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

func countRows(webhookId uint32, db *gorm.DB) (uint, error) {
	var countInt int64
	if err := db.Model(&model.Row{}).Where("webhook = ?", webhookId).Count(&countInt).Error; err != nil {
		return 0, fmt.Errorf("cannot count rows: %w", err)
	}
	return uint(countInt), nil
}

func getWebhook(hashStr string, db *gorm.DB) (*model.Webhook, error) {
	hash := model.WebhookHash(hashStr)
	webhook := model.Webhook{}

	err := db.First(&webhook, "hash = ?", hash).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if err != nil {
			return nil, &webhooks.WebhookNotFoundError{Message: fmt.Sprintf(`Webhook with hash "%s" not found.`, hashStr)}
		}
	}
	return &webhook, err
}
