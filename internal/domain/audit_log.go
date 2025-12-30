package domain

import (
	"encoding/json"
	"time"
)

type EntityType string

const (
	EntityTypeUser        EntityType = "user"
	EntityTypeTransaction EntityType = "transaction"
	EntityTypeBalance     EntityType = "balance"
)

type AuditAction string

const (
	ActionCreate AuditAction = "create"
	ActionUpdate AuditAction = "update"
	ActionDelete AuditAction = "delete"
	ActionRead   AuditAction = "read"
)

type AuditLog struct {
	ID         int64       `json:"id" db:"id"`
	EntityType EntityType  `json:"entity_type" db:"entity_type"`
	EntityID   int64       `json:"entity_id" db:"entity_id"`
	Action     AuditAction `json:"action" db:"action"`
	Details    string      `json:"details" db:"details"`
	CreatedAt  time.Time   `json:"created_at" db:"created_at"`
}

func NewAuditLog(entityType EntityType, entityID int64, action AuditAction, details interface{}) (*AuditLog, error) {
	detailsJSON, err := json.Marshal(details)
	if err != nil {
		return nil, err
	}

	return &AuditLog{
		EntityType: entityType,
		EntityID:   entityID,
		Action:     action,
		Details:    string(detailsJSON),
		CreatedAt:  time.Now(),
	}, nil
}
