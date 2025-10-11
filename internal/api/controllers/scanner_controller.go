package controllers

import (
	"ares_api/internal/api/dto"
	"ares_api/internal/common"
	"ares_api/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ScannerController struct {
	Service *services.ScannerServiceImpl
}

func NewScannerController(s *services.ScannerServiceImpl) *ScannerController {
	return &ScannerController{Service: s}
}

// @Summary Scan filesystem for Solace/ARES fragments
// @Description Recursively scans filesystem for files containing search terms
// @Tags Scanner
// @Accept json
// @Produce json
// @Param request body dto.FileScanRequest true "Scan request"
// @Success 200 {object} dto.FileScanResponse
// @Failure 400 {object} map[string]string
// @Security BearerAuth
// @Router /scanner/scan [post]
func (sc *ScannerController) ScanFiles(c *gin.Context) {
	var req dto.FileScanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.JSON(c, http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Security: Limit scan to safe directories
	if req.RootPath == "" {
		req.RootPath = "C:\\ARES_Workspace"
	}

	resp, err := sc.Service.ScanFiles(req)
	if err != nil {
		common.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	common.JSON(c, http.StatusOK, resp)
}

// @Summary Import file fragments into memory
// @Description Imports discovered fragments into memory_snapshots table
// @Tags Scanner
// @Accept json
// @Produce json
// @Param request body dto.ImportFragmentsRequest true "Import request"
// @Success 200 {object} dto.ImportFragmentsResponse
// @Failure 400 {object} map[string]string
// @Security BearerAuth
// @Router /scanner/import [post]
func (sc *ScannerController) ImportFragments(c *gin.Context) {
	var req dto.ImportFragmentsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.JSON(c, http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get userID from JWT
	userIDInterface, exists := c.Get("userID")
	if !exists {
		common.JSON(c, http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := userIDInterface.(uint)

	// Default event type
	if req.EventType == "" {
		req.EventType = "solace_fragment"
	}

	resp, err := sc.Service.ImportFragments(userID, req)
	if err != nil {
		common.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	common.JSON(c, http.StatusOK, resp)
}

// @Summary Import Solace Delta 3.1 data
// @Description Scans C:\ProgramData\Solace\State and imports all files
// @Tags Scanner
// @Accept json
// @Produce json
// @Success 200 {object} dto.ImportFragmentsResponse
// @Failure 400 {object} map[string]string
// @Security BearerAuth
// @Router /scanner/solace [post]
func (sc *ScannerController) ImportSolaceData(c *gin.Context) {
	// Get userID from JWT
	userIDInterface, exists := c.Get("userID")
	if !exists {
		common.JSON(c, http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userID := userIDInterface.(uint)

	// Call service to import Solace data
	resp, err := sc.Service.ImportSolaceData(userID)
	if err != nil {
		common.JSON(c, http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	common.JSON(c, http.StatusOK, resp)
}
