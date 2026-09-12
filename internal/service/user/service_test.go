package user

import (
	"context"
	"errors"
	"testing"

	commonerrors "common-svr/internal/common/errors"
	"common-svr/internal/model"

	"github.com/google/uuid"
)

type repositoryStub struct {
	created *model.User
}

func TestCreateRejectsInvalidEmailOutsideHTTPHandler(t *testing.T) {
	service := NewService(&repositoryStub{})

	_, err := service.Create(context.Background(), CreateInput{
		Name:  "Alice",
		Email: "not-an-email",
	})
	if !errors.Is(err, commonerrors.ErrInvalidArgument) {
		t.Fatalf("Create() error = %v, want ErrInvalidArgument", err)
	}
}

func (r *repositoryStub) Create(_ context.Context, user *model.User) error {
	r.created = user
	return nil
}

func (r *repositoryStub) GetByID(_ context.Context, id uuid.UUID) (*model.User, error) {
	return &model.User{ID: id, Name: "Existing", Email: "existing@example.com"}, nil
}

func (r *repositoryStub) List(context.Context, int, int) ([]model.User, int64, error) {
	return nil, 0, nil
}

func (r *repositoryStub) Update(context.Context, *model.User) error { return nil }
func (r *repositoryStub) Delete(context.Context, uuid.UUID) error   { return nil }

func TestCreateNormalizesInput(t *testing.T) {
	repository := &repositoryStub{}
	service := NewService(repository)

	created, err := service.Create(context.Background(), CreateInput{
		Name:  "  Alice  ",
		Email: "  ALICE@EXAMPLE.COM ",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.Name != "Alice" {
		t.Fatalf("Create() name = %q, want Alice", created.Name)
	}
	if created.Email != "alice@example.com" {
		t.Fatalf("Create() email = %q, want alice@example.com", created.Email)
	}
	if created.ID == uuid.Nil {
		t.Fatal("Create() generated a nil UUID")
	}
}
