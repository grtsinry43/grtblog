package persistence

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/grtsinry43/grtblog-v2/server/internal/domain/identity"
	"github.com/grtsinry43/grtblog-v2/server/internal/infra/persistence/model"
)

type AdminTokenRepository struct {
	db *gorm.DB
}

func NewAdminTokenRepository(db *gorm.DB) *AdminTokenRepository {
	return &AdminTokenRepository{db: db}
}

func (r *AdminTokenRepository) List(ctx context.Context, page, pageSize int) ([]identity.AdminToken, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	query := r.db.WithContext(ctx).Model(&model.AdminToken{})

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var records []model.AdminToken
	if err := query.
		Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&records).Error; err != nil {
		return nil, 0, err
	}

	items := make([]identity.AdminToken, 0, len(records))
	for _, rec := range records {
		items = append(items, mapAdminTokenToDomain(rec))
	}
	return items, total, nil
}

func (r *AdminTokenRepository) Create(ctx context.Context, token *identity.AdminToken) error {
	rec := model.AdminToken{
		Token:       token.Token,
		UserID:      token.UserID,
		Description: token.Description,
		ExpireAt:    token.ExpireAt,
	}
	if err := r.db.WithContext(ctx).Create(&rec).Error; err != nil {
		return err
	}
	token.ID = rec.ID
	token.CreatedAt = rec.CreatedAt
	token.UpdatedAt = rec.UpdatedAt
	return nil
}

func (r *AdminTokenRepository) DeleteByID(ctx context.Context, id int64) error {
	tx := r.db.WithContext(ctx).Where("id = ?", id).Delete(&model.AdminToken{})
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return identity.ErrAdminTokenNotFound
	}
	return nil
}

func (r *AdminTokenRepository) FindByToken(ctx context.Context, token string) (*identity.AdminToken, error) {
	tokenHash := HashAdminToken(token)
	var rec model.AdminToken
	if err := r.db.WithContext(ctx).Where("token = ?", tokenHash).First(&rec).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, identity.ErrAdminTokenNotFound
		}
		return nil, err
	}

	item := mapAdminTokenToDomain(rec)
	if time.Now().UTC().After(item.ExpireAt) {
		return nil, identity.ErrAdminTokenExpired
	}
	return &item, nil
}

func (r *AdminTokenRepository) IsDuplicateTokenError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "uq_admin_token_token")
}

func HashAdminToken(raw string) string {
	sum := md5.Sum([]byte(strings.TrimSpace(raw)))
	return hex.EncodeToString(sum[:])
}

func mapAdminTokenToDomain(rec model.AdminToken) identity.AdminToken {
	return identity.AdminToken{
		ID:          rec.ID,
		Token:       rec.Token,
		UserID:      rec.UserID,
		Description: rec.Description,
		ExpireAt:    rec.ExpireAt,
		CreatedAt:   rec.CreatedAt,
		UpdatedAt:   rec.UpdatedAt,
	}
}
