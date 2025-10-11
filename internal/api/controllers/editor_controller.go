package controllers

import (
	"ares_api/internal/api/dto"
	"ares_api/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type EditorController struct {
	EditorService *services.EditorServiceImpl
}

func NewEditorController(editorService *services.EditorServiceImpl) *EditorController {
	return &EditorController{
		EditorService: editorService,
	}
}

// ReadFile reads file content
// @Summary Read file
// @Tags Editor
// @Accept json
// @Produce json
// @Param request body dto.EditorFileRequest true "File path"
// @Success 200 {object} dto.EditorFileResponse
// @Router /api/v1/editor/read [post]
func (c *EditorController) ReadFile(ctx *gin.Context) {
	var req dto.EditorFileRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := c.EditorService.ReadFile(req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// SaveFile saves file content
// @Summary Save file
// @Tags Editor
// @Accept json
// @Produce json
// @Param request body dto.EditorSaveRequest true "File path and content"
// @Success 200 {object} dto.EditorSaveResponse
// @Router /api/v1/editor/save [post]
func (c *EditorController) SaveFile(ctx *gin.Context) {
	var req dto.EditorSaveRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := c.EditorService.SaveFile(req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// ListFiles lists files in directory
// @Summary List files
// @Tags Editor
// @Accept json
// @Produce json
// @Param request body dto.EditorListRequest true "Directory path"
// @Success 200 {object} dto.EditorListResponse
// @Router /api/v1/editor/list [post]
func (c *EditorController) ListFiles(ctx *gin.Context) {
	var req dto.EditorListRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := c.EditorService.ListFiles(req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// CreateFile creates a new file or directory
// @Summary Create file/directory
// @Tags Editor
// @Accept json
// @Produce json
// @Param request body dto.EditorCreateRequest true "Path and type"
// @Success 200 {object} dto.EditorSaveResponse
// @Router /api/v1/editor/create [post]
func (c *EditorController) CreateFile(ctx *gin.Context) {
	var req dto.EditorCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := c.EditorService.CreateFile(req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// DeleteFile deletes a file or directory
// @Summary Delete file/directory
// @Tags Editor
// @Accept json
// @Produce json
// @Param request body dto.EditorDeleteRequest true "Path to delete"
// @Success 200 {object} dto.EditorSaveResponse
// @Router /api/v1/editor/delete [post]
func (c *EditorController) DeleteFile(ctx *gin.Context) {
	var req dto.EditorDeleteRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := c.EditorService.DeleteFile(req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, response)
}

// RenameFile renames/moves a file or directory
// @Summary Rename/move file/directory
// @Tags Editor
// @Accept json
// @Produce json
// @Param request body dto.EditorRenameRequest true "Old and new paths"
// @Success 200 {object} dto.EditorSaveResponse
// @Router /api/v1/editor/rename [post]
func (c *EditorController) RenameFile(ctx *gin.Context) {
	var req dto.EditorRenameRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := c.EditorService.RenameFile(req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, response)
}
