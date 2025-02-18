package user

import "github.com/seanhuebl/unity-wealth/internal/services/user"

type Handler struct {
	userService *user.UserService
}

func NewHandler(userSvc *user.UserService) *Handler {
	return &Handler{
		userService: userSvc,
	}
}
