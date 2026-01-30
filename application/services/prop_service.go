package services

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	// Added missing import
	models "github.com/drama-generator/backend/domain/models"
	"github.com/drama-generator/backend/pkg/ai"
	"github.com/drama-generator/backend/pkg/config"
	"github.com/drama-generator/backend/pkg/logger"
	"github.com/drama-generator/backend/pkg/utils"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type PropService struct {
	db                     *gorm.DB
	aiService              *AIService
	taskService            *TaskService
	imageGenerationService *ImageGenerationService
	log                    *logger.Logger
	config                 *config.Config
	promptI18n             *PromptI18n
	styleConsistencyService *StyleConsistencyService
}

type PropUpsertRequest struct {
	DramaID      uint                   `json:"drama_id"`
	Name         string                 `json:"name"`
	Type         *string                `json:"type"`
	Description  *string                `json:"description"`
	Prompt       *string                `json:"prompt"`
	ImageURL     *string                `json:"image_url"`
	Attributes   map[string]interface{} `json:"attributes"`
	CharacterIDs *[]uint                `json:"character_ids"`
	SceneIDs     *[]uint                `json:"scene_ids"`
	CreatedBy    *uint                  `json:"created_by"`
}

func NewPropService(db *gorm.DB, aiService *AIService, taskService *TaskService, imageGenerationService *ImageGenerationService, log *logger.Logger, cfg *config.Config) *PropService {
	return &PropService{
		db:                     db,
		aiService:              aiService,
		taskService:            taskService,
		imageGenerationService: imageGenerationService,
		log:                    log,
		config:                 cfg,
		promptI18n:             NewPromptI18n(cfg),
		styleConsistencyService: NewStyleConsistencyService(cfg, log),
	}
}

// ListProps 获取剧本的道具列表
func (s *PropService) ListProps(dramaID uint) ([]models.Prop, error) {
	var props []models.Prop
	if err := s.db.Preload("Characters").Preload("Scenes").Where("drama_id = ?", dramaID).Find(&props).Error; err != nil {
		return nil, err
	}
	return props, nil
}

func (s *PropService) ListEpisodeProps(episodeID uint) ([]models.Prop, error) {
	var episode models.Episode
	if err := s.db.Preload("Props").Preload("Props.Characters").Preload("Props.Scenes").First(&episode, episodeID).Error; err != nil {
		return nil, err
	}
	return episode.Props, nil
}

func (s *PropService) CreatePropWithRelations(req *PropUpsertRequest) (*models.Prop, error) {
	attributesJSON := datatypes.JSON([]byte("{}"))
	if req.Attributes != nil {
		data, err := json.Marshal(req.Attributes)
		if err != nil {
			return nil, err
		}
		attributesJSON = datatypes.JSON(data)
	}

	prop := &models.Prop{
		DramaID:    req.DramaID,
		Name:       strings.TrimSpace(req.Name),
		Type:       req.Type,
		Description: req.Description,
		Prompt:     req.Prompt,
		ImageURL:   req.ImageURL,
		Attributes: attributesJSON,
		CreatedBy:  req.CreatedBy,
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(prop).Error; err != nil {
			return err
		}

		if req.CharacterIDs != nil {
			var characters []models.Character
			if len(*req.CharacterIDs) > 0 {
				if err := tx.Where("id IN ?", *req.CharacterIDs).Find(&characters).Error; err != nil {
					return err
				}
			}
			if err := tx.Model(prop).Association("Characters").Replace(characters); err != nil {
				return err
			}
		}

		if req.SceneIDs != nil {
			var scenes []models.Scene
			if len(*req.SceneIDs) > 0 {
				if err := tx.Where("id IN ?", *req.SceneIDs).Find(&scenes).Error; err != nil {
					return err
				}
			}
			if err := tx.Model(prop).Association("Scenes").Replace(scenes); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return prop, nil
}

func (s *PropService) UpdatePropWithRelations(id uint, req *PropUpsertRequest) (*models.Prop, error) {
	var prop models.Prop
	if err := s.db.First(&prop, id).Error; err != nil {
		return nil, err
	}

	updates := map[string]interface{}{}
	if strings.TrimSpace(req.Name) != "" {
		updates["name"] = strings.TrimSpace(req.Name)
	}
	if req.Type != nil {
		updates["type"] = req.Type
	}
	if req.Description != nil {
		updates["description"] = req.Description
	}
	if req.Prompt != nil {
		updates["prompt"] = req.Prompt
	}
	if req.ImageURL != nil {
		updates["image_url"] = req.ImageURL
	}
	if req.Attributes != nil {
		data, err := json.Marshal(req.Attributes)
		if err != nil {
			return nil, err
		}
		updates["attributes"] = datatypes.JSON(data)
	}
	if req.CreatedBy != nil {
		updates["created_by"] = req.CreatedBy
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if len(updates) > 0 {
			if err := tx.Model(&prop).Updates(updates).Error; err != nil {
				return err
			}
		}

		if req.CharacterIDs != nil {
			var characters []models.Character
			if len(*req.CharacterIDs) > 0 {
				if err := tx.Where("id IN ?", *req.CharacterIDs).Find(&characters).Error; err != nil {
					return err
				}
			}
			if err := tx.Model(&prop).Association("Characters").Replace(characters); err != nil {
				return err
			}
		}

		if req.SceneIDs != nil {
			var scenes []models.Scene
			if len(*req.SceneIDs) > 0 {
				if err := tx.Where("id IN ?", *req.SceneIDs).Find(&scenes).Error; err != nil {
					return err
				}
			}
			if err := tx.Model(&prop).Association("Scenes").Replace(scenes); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	if err := s.db.Preload("Characters").Preload("Scenes").First(&prop, prop.ID).Error; err != nil {
		return nil, err
	}

	return &prop, nil
}

// DeleteProp 删除道具
func (s *PropService) DeleteProp(id uint) error {
	return s.db.Delete(&models.Prop{}, id).Error
}

func (s *PropService) AddPropToLibrary(propID uint, userID uint, permission string) (*models.PropLibrary, error) {
	if strings.TrimSpace(permission) == "" {
		permission = "read"
	}
	item := &models.PropLibrary{
		PropID:     propID,
		UserID:     userID,
		Permission: permission,
	}
	if err := s.db.Create(item).Error; err != nil {
		return nil, err
	}
	return item, nil
}

func (s *PropService) ListPropLibrary(userID uint) ([]models.PropLibrary, error) {
	var items []models.PropLibrary
	if err := s.db.Preload("Prop").Where("user_id = ?", userID).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (s *PropService) UpdatePropLibraryPermission(id uint, permission string) error {
	if strings.TrimSpace(permission) == "" {
		return fmt.Errorf("permission is required")
	}
	return s.db.Model(&models.PropLibrary{}).Where("id = ?", id).Update("permission", permission).Error
}

func (s *PropService) RemovePropFromLibrary(id uint) error {
	return s.db.Delete(&models.PropLibrary{}, id).Error
}

// ExtractPropsFromScript 从剧本提取道具（异步）
func (s *PropService) ExtractPropsFromScript(episodeID uint) (string, error) {
	var episode models.Episode
	if err := s.db.First(&episode, episodeID).Error; err != nil {
		return "", fmt.Errorf("episode not found: %w", err)
	}

	if episode.ScriptContent == nil || strings.TrimSpace(*episode.ScriptContent) == "" {
		return "", fmt.Errorf("剧本内容为空")
	}

	task, err := s.taskService.CreateTask("prop_extraction", fmt.Sprintf("%d", episodeID))
	if err != nil {
		return "", err
	}

	go s.processPropExtraction(task.ID, episode)

	return task.ID, nil
}

func (s *PropService) processPropExtraction(taskID string, episode models.Episode) {
	s.taskService.UpdateTaskStatus(taskID, "processing", 0, "正在分析剧本...")

	script := ""
	if episode.ScriptContent != nil {
		script = *episode.ScriptContent
	}

	promptTemplate := s.promptI18n.GetPropExtractionPrompt()
	prompt := fmt.Sprintf(promptTemplate, script)

	response, err := s.aiService.GenerateText(prompt, "", ai.WithMaxTokens(2000))
	if err != nil {
		s.taskService.UpdateTaskError(taskID, err)
		return
	}

	var extractedProps []struct {
		Name        string `json:"name"`
		Type        string `json:"type"`
		Description string `json:"description"`
		ImagePrompt string `json:"image_prompt"`
	}

	if err := utils.SafeParseAIJSON(response, &extractedProps); err != nil {
		s.taskService.UpdateTaskError(taskID, fmt.Errorf("解析AI结果失败: %w", err))
		return
	}

	s.taskService.UpdateTaskStatus(taskID, "processing", 50, "正在保存道具...")

	var createdProps []models.Prop
	for _, p := range extractedProps {
		var count int64
		s.db.Model(&models.Prop{}).Where("drama_id = ? AND name = ?", episode.DramaID, p.Name).Count(&count)
		if count > 0 {
			var existingProp models.Prop
			if err := s.db.Where("drama_id = ? AND name = ?", episode.DramaID, p.Name).First(&existingProp).Error; err == nil {
				_ = s.db.Model(&episode).Association("Props").Append(&existingProp)
			}
			continue
		}

		prop := models.Prop{
			DramaID:     episode.DramaID,
			Name:        p.Name,
			Type:        &p.Type,
			Description: &p.Description,
			Prompt:      &p.ImagePrompt,
		}
		if err := s.db.Create(&prop).Error; err == nil {
			createdProps = append(createdProps, prop)
			_ = s.db.Model(&episode).Association("Props").Append(&prop)
		}
	}

	s.taskService.UpdateTaskResult(taskID, createdProps)
}

func (s *PropService) GeneratePropImage(propID uint) (string, error) {
	// 1. 获取道具信息
	var prop models.Prop
	if err := s.db.First(&prop, propID).Error; err != nil {
		return "", err
	}

	if prop.Prompt == nil || *prop.Prompt == "" {
		return "", fmt.Errorf("道具没有图片提示词")
	}

	// 2. 创建任务
	task, err := s.taskService.CreateTask("prop_image_generation", fmt.Sprintf("%d", propID))
	if err != nil {
		return "", err
	}

	go s.processPropImageGeneration(task.ID, prop)
	return task.ID, nil
}

func (s *PropService) processPropImageGeneration(taskID string, prop models.Prop) {
	s.taskService.UpdateTaskStatus(taskID, "processing", 0, "正在生成图片...")

	// 准备生成参数
	imageStyle := s.config.Style.DefaultStyle
	targetStyle := ""
	referenceWork := ""
	if s.config != nil && s.config.Style.DefaultPropStyle != "" {
		imageStyle += ", " + s.config.Style.DefaultPropStyle
	}

	// Fetch Drama to get StylePrompt
	var drama models.Drama
	if err := s.db.First(&drama, prop.DramaID).Error; err == nil {
		if drama.StylePrompt != nil && *drama.StylePrompt != "" {
			imageStyle += ", " + *drama.StylePrompt
			targetStyle = *drama.StylePrompt
		} else if drama.Style != "" && drama.Style != "realistic" {
			targetStyle = drama.Style
		}
		if drama.ReferenceWork != nil && *drama.ReferenceWork != "" {
			referenceWork = *drama.ReferenceWork
		}
	}

	imageSize := "1024x1024"
	if s.config != nil && s.config.Style.DefaultImageSize != "" {
		imageSize = s.config.Style.DefaultImageSize
	}

	prompt := *prop.Prompt
	if s.styleConsistencyService != nil {
		normalizedPrompt, violations := s.styleConsistencyService.NormalizeAndValidatePrompt(prompt, targetStyle, referenceWork)
		if len(violations) > 0 {
			s.log.Infow("Prop prompt normalized by style consistency", "prop_id", prop.ID, "violations", violations)
		}
		prompt = normalizedPrompt
	}

	// 创建生成请求
	req := &GenerateImageRequest{
		DramaID:   fmt.Sprintf("%d", prop.DramaID),
		PropID:    &prop.ID,
		ImageType: string(models.ImageTypeProp),
		Prompt:    prompt,
		Size:      imageSize,
		Style:     &imageStyle,
		Provider:  s.config.AI.DefaultImageProvider, // 使用默认配置
	}

	// 增加默认比例（如果有配置）
	// 注意：GenerateImageRequest 最好增加 Ratio 字段，或者在 PropService 这里拼接到 prompt
	// 目前 ImageGenerationService 支持 Width/Height 和 Size
	// 如果配置了 Ratio，可能需要转换为 Size 或 Model 参数
	// 简单起见，如果 Prompt 中需要 Ratio，可以追加
	if s.config != nil && s.config.Style.DefaultPropRatio != "" {
		// 这里暂不处理，因为 ImageGenerationService 主要靠 Size 控制
		// 如果需要，可以调整 prompt
		// req.Prompt += ", ratio: " + s.config.Style.DefaultPropRatio
	}

	// 调用 ImageGenerationService
	imageGen, err := s.imageGenerationService.GenerateImage(req)
	if err != nil {
		s.taskService.UpdateTaskError(taskID, err)
		return
	}

	// 轮询 ImageGeneration 状态直到完成
	maxAttempts := 60
	pollInterval := 2 * time.Second

	for i := 0; i < maxAttempts; i++ {
		time.Sleep(pollInterval)

		// 重新加载 imageGen
		var currentImageGen models.ImageGeneration
		if err := s.db.First(&currentImageGen, imageGen.ID).Error; err != nil {
			s.log.Errorw("Failed to poll image generation", "error", err, "id", imageGen.ID)
			continue
		}

		if currentImageGen.Status == models.ImageStatusCompleted {
			if currentImageGen.ImageURL != nil {
				// 任务成功
				// ImageGenerationService 已经更新了 Prop.ImageURL，这里只需要更新 TaskService
				s.taskService.UpdateTaskResult(taskID, map[string]string{"image_url": *currentImageGen.ImageURL})
				return
			}
		} else if currentImageGen.Status == models.ImageStatusFailed {
			errMsg := "图片生成失败"
			if currentImageGen.ErrorMsg != nil {
				errMsg = *currentImageGen.ErrorMsg
			}
			s.taskService.UpdateTaskError(taskID, fmt.Errorf(errMsg))
			return
		}

		// 更新进度（可选）
		s.taskService.UpdateTaskStatus(taskID, "processing", 10+i, "正在生成图片...")
	}

	s.taskService.UpdateTaskError(taskID, fmt.Errorf("生成超时"))
}

// AssociatePropsWithStoryboard 关联道具到分镜
func (s *PropService) AssociatePropsWithStoryboard(storyboardID uint, propIDs []uint) error {
	var storyboard models.Storyboard
	if err := s.db.First(&storyboard, storyboardID).Error; err != nil {
		return err
	}

	var props []models.Prop
	if len(propIDs) > 0 {
		if err := s.db.Where("id IN ?", propIDs).Find(&props).Error; err != nil {
			return err
		}
	}

	return s.db.Model(&storyboard).Association("Props").Replace(props)
}
