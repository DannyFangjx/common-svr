package user

import (
	"context"
	"fmt"
	"net/mail"
	"strings"
	"unicode/utf8"

	commonerrors "common-svr/internal/common/errors"
	"common-svr/internal/model"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	List(ctx context.Context, limit, offset int) ([]model.User, int64, error)
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type UserService interface {
	Create(ctx context.Context, input CreateInput) (*model.User, error)
	Get(ctx context.Context, id uuid.UUID) (*model.User, error)
	List(ctx context.Context, limit, offset int) ([]model.User, int64, error)
	Update(ctx context.Context, id uuid.UUID, input UpdateInput) (*model.User, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type CreateInput struct {
	Name  string
	Email string
}

type UpdateInput struct {
	Name  *string
	Email *string
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*model.User, error) {
	name := strings.TrimSpace(input.Name)
	email := normalizeEmail(input.Email)
	if err := validateName(name); err != nil {
		return nil, err
	}
	if err := validateEmail(email); err != nil {
		return nil, err
	}

	user := &model.User{ID: uuid.New(), Name: name, Email: email}
	if err := s.repository.Create(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (*model.User, error) {
	return s.repository.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context, limit, offset int) ([]model.User, int64, error) {
	if limit < 1 || limit > 100 || offset < 0 {
		return nil, 0, fmt.Errorf("%w: limit must be 1-100 and offset cannot be negative", commonerrors.ErrInvalidArgument)
	}
	return s.repository.List(ctx, limit, offset)
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, input UpdateInput) (*model.User, error) {
	if input.Name == nil && input.Email == nil {
		return nil, fmt.Errorf("%w: at least one field is required", commonerrors.ErrInvalidArgument)
	}

	user, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if err := validateName(name); err != nil {
			return nil, err
		}
		user.Name = name
	}
	if input.Email != nil {
		email := normalizeEmail(*input.Email)
		if err := validateEmail(email); err != nil {
			return nil, err
		}
		user.Email = email
	}

	if err := s.repository.Update(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repository.Delete(ctx, id)
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func validateName(name string) error {
	if name == "" {
		return fmt.Errorf("%w: name is required", commonerrors.ErrInvalidArgument)
	}
	if utf8.RuneCountInString(name) > 100 {
		return fmt.Errorf("%w: name is too long", commonerrors.ErrInvalidArgument)
	}
	return nil
}

func validateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("%w: email is required", commonerrors.ErrInvalidArgument)
	}
	if len(email) > 320 {
		return fmt.Errorf("%w: email is too long", commonerrors.ErrInvalidArgument)
	}
	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email {
		return fmt.Errorf("%w: email format is invalid", commonerrors.ErrInvalidArgument)
	}
	return nil
}
