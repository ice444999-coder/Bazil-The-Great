package controllers

import (
	"ares_api/internal/api/dto"
	"ares_api/internal/common"
	interfaces "ares_api/internal/interfaces/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	Service interfaces.UserService
	LedgerService interfaces.LedgerService
}

func NewUserController(service interfaces.UserService , ledgerService interfaces.LedgerService) *UserController {
	return &UserController{Service: service , LedgerService: ledgerService}
}

// @Summary Signup
// @Description Register a new user
// @Tags Auth
// @Accept  json
// @Produce  json
// @Param   user body dto.SignupRequest true "Signup Data"
// @Success 200 {object} dto.SignupResponse
// @Failure 400 {object} map[string]string
// @Router /users/signup [post]
func (uc *UserController) Signup(c *gin.Context) {
	var req dto.SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.JSON(c, http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := uc.Service.Signup(req.Username, req.Email, req.Password); err != nil {
		common.JSON(c, http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	_ = uc.LedgerService.Append(0, "signup", "New user signed up: "+req.Username)
	common.JSON(c, http.StatusOK, dto.SignupResponse{Message: "Signup successful"})
}

// @Summary Login
// @Description Authenticate a user and return JWT
// @Tags Auth
// @Accept  json
// @Produce  json
// @Param   user body dto.LoginRequest true "Login Data"
// @Success 200 {object} dto.LoginResponse
// @Failure 401 {object} map[string]string
// @Router /users/login [post]
func (uc *UserController) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.JSON(c, http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userid ,accessToken, refreshToken, err := uc.Service.Login(req.Username, req.Password)
	if err != nil {
		common.JSON(c, http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	_ = uc.LedgerService.Append(userid, "login", "User logged in: "+req.Username)
	common.JSON(c, http.StatusOK, dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

// @Summary Refresh Token
// @Description Generate a new access token using refresh token
// @Tags Auth
// @Accept  json
// @Produce  json
// @Param   token body dto.RefreshRequest true "Refresh Token"
// @Success 200 {object} dto.RefreshResponse
// @Failure 401 {object} map[string]string
// @Router /users/refresh [post]
func (uc *UserController) RefreshToken(c *gin.Context) {
	var req dto.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.JSON(c, http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newAccessToken, err := uc.Service.Refresh(req.RefreshToken)
	if err != nil {
		common.JSON(c, http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
     _ = uc.LedgerService.Append(0, "token_refresh", "Access token refreshed")
	common.JSON(c, http.StatusOK, dto.RefreshResponse{
		AccessToken: newAccessToken,
	})
}
