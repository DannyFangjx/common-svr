package user

type CreateRequest struct {
	Name  string `json:"name" binding:"required,max=100"`
	Email string `json:"email" binding:"required,email,max=320"`
}

type GetRequest struct {
	UserID string `form:"user_id" binding:"required,uuid"`
}

type ListRequest struct {
	Limit  int `form:"limit,default=20" binding:"min=1,max=100"`
	Offset int `form:"offset,default=0" binding:"min=0"`
}

type UpdateRequest struct {
	UserID string  `json:"user_id" binding:"required,uuid"`
	Name   *string `json:"name" binding:"omitempty,min=1,max=100"`
	Email  *string `json:"email" binding:"omitempty,email,max=320"`
}

type DeleteRequest struct {
	UserID string `json:"user_id" binding:"required,uuid"`
}
