package testhelpers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/seanhuebl/unity-wealth/internal/constants"
)

func CheckForUserIDIssues(tcName string, userID uuid.UUID, c *gin.Context) {
	if tcName == "unauthorized: user ID not UUID" {
		c.Set(string(constants.UserIDKey), "userID")
	} else {
		c.Set(string(constants.UserIDKey), userID)
	}
}
