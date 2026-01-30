package services

import (
	"strings"
	"testing"

	"github.com/drama-generator/backend/pkg/config"
)

func TestNormalizePromptByStyle_AppendsWithoutReplacing(t *testing.T) {
	prompt := "masterpiece, japanese anime style, warm light"
	style := "Chinese Animation Style"
	referenceWork := "Test Ref"

	output := normalizePromptByStyle(prompt, style, referenceWork)

	if !strings.Contains(strings.ToLower(output), "japanese anime style") {
		t.Fatalf("expected to keep existing style segment, got: %s", output)
	}
	if !strings.Contains(output, "Chinese Animation Style") {
		t.Fatalf("expected to append target style, got: %s", output)
	}
	if !strings.Contains(output, "style reference: Test Ref") {
		t.Fatalf("expected to append reference work, got: %s", output)
	}
}

func TestNormalizePromptByStyle_DedupAndNormalizeReference(t *testing.T) {
	prompt := "best quality, chinese animation style, inspired by Test Ref"
	style := "Chinese Animation Style"
	referenceWork := "Test Ref"

	output := normalizePromptByStyle(prompt, style, referenceWork)

	if strings.Count(output, "style reference: Test Ref") != 1 {
		t.Fatalf("expected single reference marker, got: %s", output)
	}
	if !strings.Contains(strings.ToLower(output), "chinese animation style") {
		t.Fatalf("expected to keep existing style text, got: %s", output)
	}
	if strings.Count(strings.ToLower(output), strings.ToLower(style)) > 1 {
		t.Fatalf("expected target style to avoid duplication, got: %s", output)
	}
}

func TestStyleConsistency_RemovesConflictingStyle(t *testing.T) {
	service := NewStyleConsistencyService(&config.Config{}, nil)
	prompt := "best quality, japanese anime style, warm light"
	style := "Chinese Animation Style"

	output, violations := service.NormalizeAndValidatePrompt(prompt, style, "")

	if strings.Contains(strings.ToLower(output), "japanese anime style") {
		t.Fatalf("expected to remove conflicting style, got: %s", output)
	}
	if !strings.Contains(output, style) {
		t.Fatalf("expected to append target style, got: %s", output)
	}
	if len(violations) == 0 {
		t.Fatalf("expected violations for conflicting style, got none")
	}
}

func TestStyleConsistency_NormalizesReferenceWork(t *testing.T) {
	service := NewStyleConsistencyService(&config.Config{}, nil)
	prompt := "best quality, style reference: Other Ref, cinematic lighting"
	style := "Chinese Animation Style"
	referenceWork := "Target Ref"

	output, _ := service.NormalizeAndValidatePrompt(prompt, style, referenceWork)

	if strings.Count(output, "style reference: Target Ref") != 1 {
		t.Fatalf("expected single normalized reference, got: %s", output)
	}
	if !strings.Contains(output, style) {
		t.Fatalf("expected target style to be present, got: %s", output)
	}
}
