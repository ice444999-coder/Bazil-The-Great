package controllers

import (
	"ares_api/internal/api/dto"
	"ares_api/internal/common"
	service "ares_api/internal/interfaces/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SettingsController struct {
	Service service.SettingsService
	LedgerService service.LedgerService
}

func NewSettingsController(s service.SettingsService , l service.LedgerService) *SettingsController {
	return &SettingsController{Service: s , LedgerService: l}
}

// @Summary Save API Key
// @Tags Settings
// @Accept  json
// @Produce  json
// @Param   request body dto.APIKeyRequest true "API Key"
// @Success 200 {object} dto.APIKeyResponse
// @Security BearerAuth
// @Router /settings/apikey [post]
func (sc *SettingsController) SaveAPIKey(c *gin.Context) {
	var req dto.APIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.JSON(c, http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetUint("userID")
	if err := sc.Service.SaveAPIKey(userID, req.APIKey); err != nil {
		common.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	_ = sc.LedgerService.Append(userID, "settings", "Saved new API key")

	common.JSON(c, http.StatusOK, dto.APIKeyResponse{Message: "API key saved successfully"})
}

