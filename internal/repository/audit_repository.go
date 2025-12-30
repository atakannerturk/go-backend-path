package repository

import (
	"context"

	"github.com/atakannerturk/go-backend-path/internal/domain"
)

type AuditRepository interface {
	Create(ctx context.Context, log *domain.AuditLog) error
	GetByEntityID(ctx context.Context, entityType domain.EntityType, entityID int64, limit, offset int) ([]*domain.AuditLog, error)
	List(ctx context.Context, limit, offset int) ([]*domain.AuditLog, error)
}
