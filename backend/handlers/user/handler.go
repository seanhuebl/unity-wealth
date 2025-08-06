package user

type Handler struct {
	userService UserService
}

func NewHandler(userSvc UserService) *Handler {
	return &Handler{
		userService: userSvc,
	}
}
