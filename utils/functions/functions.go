package functions

import (
	"digital-book-lending/utils"
	"digital-book-lending/utils/response"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func ValidateUUID(ctx *gin.Context, logPrefix string, logID uuid.UUID) (string, error) {
	id := ctx.Param("id")
	if id == "" {
		utils.WriteLog(utils.LogLevelError, fmt.Sprintf("%s; Missing ID in path", logPrefix))
		res := response.Response(http.StatusBadRequest, http.StatusText(http.StatusBadRequest), logID, nil)
		res.Error = "ID parameter is required"
		ctx.JSON(http.StatusBadRequest, res)
		return "", fmt.Errorf("missing ID")
	}

	if _, err := uuid.Parse(id); err != nil {
		utils.WriteLog(utils.LogLevelError, fmt.Sprintf("%s; Invalid ID: '%s'", logPrefix, id))
		res := response.Response(http.StatusBadRequest, http.StatusText(http.StatusBadRequest), logID, nil)
		res.Errors = response.Errors{Code: http.StatusBadRequest, Message: "ID must be a valid UUID"}
		ctx.JSON(http.StatusBadRequest, res)
		return "", fmt.Errorf("invalid UUID")
	}

	return id, nil
}
