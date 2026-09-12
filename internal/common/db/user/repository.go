package user

import (
	"context"
	"errors"

	commonerrors "common-svr/internal/common/errors"
	"common-svr/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, user *model.User) error {
	return translateError(r.db.WithContext(ctx).Create(user).Error)
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var user model.User
	if err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error; err != nil {
		return nil, translateError(err)
	}
	return &user, nil
}

func (r *Repository) List(ctx context.Context, limit, offset int) ([]model.User, int64, error) {
	var (
		users []model.User
		total int64
	)
	db := r.db.WithContext(ctx).Model(&model.User{})
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, translateError(err)
	}
	if err := db.Order("created_at DESC").Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		return nil, 0, translateError(err)
	}
	return users, total, nil
}

func (r *Repository) Update(ctx context.Context, user *model.User) error {
	result := r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ?", user.ID).
		Updates(map[string]any{"name": user.Name, "email": user.Email})
	if result.Error != nil {
		return translateError(result.Error)
	}
	if result.RowsAffected == 0 {
		return commonerrors.ErrNotFound
	}
	return translateError(r.db.WithContext(ctx).First(user, "id = ?", user.ID).Error)
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&model.User{}, "id = ?", id)
	if result.Error != nil {
		return translateError(result.Error)
	}
	if result.RowsAffected == 0 {
		return commonerrors.ErrNotFound
	}
	return nil
}

func translateError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, gorm.ErrRecordNotFound):
		return commonerrors.ErrNotFound
	case errors.Is(err, gorm.ErrDuplicatedKey):
		return commonerrors.ErrConflict
	default:
		return err
	}
}
