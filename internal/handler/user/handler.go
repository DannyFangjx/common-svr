package user

import (
	"context"
	"fmt"
	"net/http"

	commonerrors "common-svr/internal/common/errors"
	"common-svr/internal/common/response"
	"common-svr/internal/model"
	userservice "common-svr/internal/service/user"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type userService interface {
	Create(ctx context.Context, input userservice.CreateInput) (*model.User, error)
	Get(ctx context.Context, id uuid.UUID) (*model.User, error)
	List(ctx context.Context, limit, offset int) ([]model.User, int64, error)
	Update(ctx context.Context, id uuid.UUID, input userservice.UpdateInput) (*model.User, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type Handler struct {
	service userService
}

func NewHandler(service userService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Create(c *gin.Context) {
	var request CreateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Failure(c, fmt.Errorf("%w: %v", commonerrors.ErrInvalidArgument, err))
		return
	}

	created, err := h.service.Create(c.Request.Context(), userservice.CreateInput{
		Name:  request.Name,
		Email: request.Email,
	})
	if err != nil {
		response.Failure(c, err)
		return
	}
	response.Success(c, http.StatusCreated, CreateResponse{User: toResponse(created)})
}

func (h *Handler) Get(c *gin.Context) {
	var request GetRequest
	if err := c.ShouldBindQuery(&request); err != nil {
		response.Failure(c, fmt.Errorf("%w: %v", commonerrors.ErrInvalidArgument, err))
		return
	}
	id := uuid.MustParse(request.UserID)
	user, err := h.service.Get(c.Request.Context(), id)
	if err != nil {
		response.Failure(c, err)
		return
	}
	response.Success(c, http.StatusOK, GetResponse{User: toResponse(user)})
}

func (h *Handler) List(c *gin.Context) {
	var request ListRequest
	if err := c.ShouldBindQuery(&request); err != nil {
		response.Failure(c, fmt.Errorf("%w: %v", commonerrors.ErrInvalidArgument, err))
		return
	}

	users, total, err := h.service.List(c.Request.Context(), request.Limit, request.Offset)
	if err != nil {
		response.Failure(c, err)
		return
	}
	items := make([]UserResponse, 0, len(users))
	for i := range users {
		items = append(items, toResponse(&users[i]))
	}
	response.Success(c, http.StatusOK, ListResponse{
		Items:  items,
		Total:  total,
		Limit:  request.Limit,
		Offset: request.Offset,
	})
}

func (h *Handler) Update(c *gin.Context) {
	var request UpdateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Failure(c, fmt.Errorf("%w: %v", commonerrors.ErrInvalidArgument, err))
		return
	}
	id := uuid.MustParse(request.UserID)
	updated, err := h.service.Update(c.Request.Context(), id, userservice.UpdateInput{
		Name:  request.Name,
		Email: request.Email,
	})
	if err != nil {
		response.Failure(c, err)
		return
	}
	response.Success(c, http.StatusOK, UpdateResponse{User: toResponse(updated)})
}

func (h *Handler) Delete(c *gin.Context) {
	var request DeleteRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Failure(c, fmt.Errorf("%w: %v", commonerrors.ErrInvalidArgument, err))
		return
	}
	id := uuid.MustParse(request.UserID)
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		response.Failure(c, err)
		return
	}
	response.Success(c, http.StatusOK, DeleteResponse{UserID: id.String()})
}
