package services

import (
	"strings"

	"github.com/drama-generator/backend/pkg/config"
	"github.com/drama-generator/backend/pkg/logger"
)

// StyleConsistencyService 风格一致性校验服务（统一校验与自动修正）
// 目标：确保所有提示词严格遵循项目创建阶段的视觉风格与参考作品规范
type StyleConsistencyService struct {
	config *config.Config
	log    *logger.Logger
}

// NewStyleConsistencyService 创建风格一致性校验服务
func NewStyleConsistencyService(cfg *config.Config, log *logger.Logger) *StyleConsistencyService {
	return &StyleConsistencyService{config: cfg, log: log}
}

// NormalizeAndValidatePrompt 归一化并校验提示词风格一致性
// 输入：原始提示词、项目风格、参考作品
// 输出：修正后的提示词、违规项列表（如有）
func (s *StyleConsistencyService) NormalizeAndValidatePrompt(prompt string, style string, referenceWork string) (string, []string) {
	violations := []string{}

	raw := strings.TrimSpace(prompt)
	if raw == "" {
		return raw, violations
	}

	// 规范标点并切分片段
	replacer := strings.NewReplacer("，", ",", "\n", ",", "、", ",")
	normalized := replacer.Replace(raw)
	segments := strings.Split(normalized, ",")

	targetStyle := strings.TrimSpace(style)
	targetRef := strings.TrimSpace(referenceWork)

	// 常见风格关键词（用于冲突检测）
	styleKeywords := []string{
		"modern japanese anime style", "modern japanese anime", "japanese anime", "anime style", "anime", "日本动漫", "日式动漫", "日本动画",
		"studio ghibli", "ghibli", "makoto shinkai", "shinkai", "新海诚", "吉卜力", "宫崎骏",
		"american animation", "western animation", "pixar", "disney", "欧美动画", "美式动画",
		"european animation", "欧式动画",
		"chinese animation", "guoman", "国漫", "中国动画", "国风", "国漫风格",
		"cel-shaded", "cel shaded", "cel shading", "cel-shading", "赛璐璐",
		"light novel cover", "轻小说封面",
		"magical girl", "魔法少女",
		"chibi", "q版", "q 版",
	}

	// 参考作品标记关键词
	refMarkers := []string{"style reference", "参考作品", "reference work"}

	// 白底与构图冲突关键词（非保护段落清理）
	whiteBgConflict := []string{
		"dark background", "black background", "gray background", "grey background", "gradient background",
		"vignette", "shadow", "floor", "ground", "scenery", "environment",
	}

	result := make([]string, 0, len(segments))
	seen := map[string]struct{}{}
	hasStyle := false
	hasReference := false

	for _, seg := range segments {
		seg = strings.TrimSpace(seg)
		if seg == "" {
			continue
		}
		lower := strings.ToLower(seg)

		// 参考作品归一化：只允许项目设定的参考作品
		if targetRef != "" {
			isRef := false
			for _, marker := range refMarkers {
				if strings.Contains(lower, marker) {
					isRef = true
					break
				}
			}
			if isRef {
				// 如果该片段不是指向目标参考作品，则替换为统一标记
				if !strings.Contains(lower, strings.ToLower(targetRef)) {
					violations = append(violations, "检测到非项目参考作品，已替换为项目参考作品")
				}
				seg = "style reference: " + targetRef
				lower = strings.ToLower(seg)
				hasReference = true
			}
		}

		// 风格冲突清理：存在明显与目标风格不一致的关键词时移除该片段
		if targetStyle != "" {
			// 如果片段中包含其他风格关键词且不包含目标风格，则判定为冲突
			containsOtherStyle := false
			for _, kw := range styleKeywords {
				if strings.Contains(lower, kw) && !strings.Contains(lower, strings.ToLower(targetStyle)) {
					containsOtherStyle = true
					break
				}
			}
			if containsOtherStyle {
				violations = append(violations, "检测到与项目风格不符的风格关键词，已移除")
				continue
			}
		}

		// 白底构图冲突清理（仅在检测到白底关键词时）
		// 白底保护关键词集合（前端也有同逻辑，这里做服务端兜底）
		whiteProtected := []string{"white background", "simple background", "studio lighting"}
		hasWhiteProtected := false
		for _, kw := range whiteProtected {
			if strings.Contains(lower, kw) {
				hasWhiteProtected = true
				break
			}
		}
		if hasWhiteProtected {
			for _, conflict := range whiteBgConflict {
				if strings.Contains(lower, conflict) {
					violations = append(violations, "检测到与白底图冲突的关键词，已移除")
					continue // 丢弃该片段
				}
			}
		}

		// 记录风格与参考作品存在性
		if targetStyle != "" && strings.Contains(lower, strings.ToLower(targetStyle)) {
			hasStyle = true
		}
		if targetRef != "" && strings.Contains(lower, strings.ToLower(targetRef)) {
			hasReference = true
		}

		key := strings.ToLower(seg)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, seg)
	}

	// 追加缺失的风格和参考作品
	if targetStyle != "" && !hasStyle {
		result = append(result, targetStyle)
	}
	if targetRef != "" && !hasReference {
		result = append(result, "style reference: "+targetRef)
	}

	final := strings.Join(result, ", ")
	return final, violations
}
