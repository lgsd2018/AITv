package handlers

import (
	"github.com/drama-generator/backend/application/services"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/drama-generator/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type AIHandler struct {
	aiService *services.AIService
	log       *logger.Logger
}

func NewAIHandler(aiService *services.AIService, log *logger.Logger) *AIHandler {
	return &AIHandler{
		aiService: aiService,
		log:       log,
	}
}

// GeneratePromptFromImageRequest defines the request body for generating prompt from image
type GeneratePromptFromImageRequest struct {
	ImageURL string `json:"image_url" binding:"required"`
}

// GeneratePromptFromImage handles the request to generate prompt from image
func (h *AIHandler) GeneratePromptFromImage(c *gin.Context) {
	var req GeneratePromptFromImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	prompt, err := h.aiService.GeneratePromptFromImage(req.ImageURL)
	if err != nil {
		h.log.Errorw("Failed to generate prompt from image", "error", err, "image_url", req.ImageURL)
		response.InternalError(c, "Failed to generate prompt: "+err.Error())
		return
	}

	response.Success(c, gin.H{"prompt": prompt})
}
