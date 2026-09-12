package user

import (
	"time"

	"common-svr/internal/model"
)

type UserResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateResponse struct {
	User UserResponse `json:"user"`
}

type GetResponse struct {
	User UserResponse `json:"user"`
}

type ListResponse struct {
	Items  []UserResponse `json:"items"`
	Total  int64          `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}

type UpdateResponse struct {
	User UserResponse `json:"user"`
}

type DeleteResponse struct {
	UserID string `json:"user_id"`
}

func toResponse(user *model.User) UserResponse {
	return UserResponse{
		ID:        user.ID.String(),
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}
