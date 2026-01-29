package handlers

import (
	"github.com/drama-generator/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

// GenerateCharacterImage AI生成角色形象
func (h *CharacterLibraryHandler) GenerateCharacterImage(c *gin.Context) {

	characterID := c.Param("id")

	// 获取请求体中的model和style参数
	var req struct {
		Model           string `json:"model"`
		Style           string `json:"style"`
		Size            string `json:"size"`
		ReferenceWork   string `json:"reference_work"`
		WhiteBackground bool   `json:"white_background"`
	}
	c.ShouldBindJSON(&req)

	imageGen, err := h.libraryService.GenerateCharacterImage(characterID, h.imageService, req.Model, req.Style, req.Size, req.ReferenceWork, req.WhiteBackground)
	if err != nil {
		if err.Error() == "character not found" {
			response.NotFound(c, "角色不存在")
			return
		}
		if err.Error() == "unauthorized" {
			response.Forbidden(c, "无权限")
			return
		}
		h.log.Errorw("Failed to generate character image", "error", err)
		response.InternalError(c, "生成失败")
		return
	}

	response.Success(c, gin.H{
		"message":          "角色图片生成已启动",
		"image_generation": imageGen,
	})
}
