package user

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/seanhuebl/unity-wealth/internal/services/auth"
	"github.com/seanhuebl/unity-wealth/internal/services/user"
)

func (h *Handler) SignUp(ctx *gin.Context) {
	var input user.SignUpInput

	if err := ctx.ShouldBindJSON(&input); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"data": gin.H{
				"error": "invalid request",
			},
		})
		return
	}
	if err := h.userService.SignUp(ctx, input); err != nil {
		switch {
		case errors.Is(err, auth.ErrInvalidEmail):
			ctx.JSON(http.StatusBadRequest, gin.H{
				"data": gin.H{
					"error": "invalid email",
				},
			})
		case errors.Is(err, auth.ErrInvalidPassword):
			ctx.JSON(http.StatusBadRequest, gin.H{
				"data": gin.H{
					"error": "invalid password",
				},
			})
		default:
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"data": gin.H{
					"error": "internal server error",
				},
			})
		}
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"data": gin.H{
			"message": "sign up successful!",
			"email":   input.Email,
		},
	})
}
