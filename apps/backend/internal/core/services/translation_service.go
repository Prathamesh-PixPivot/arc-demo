package services

import (
	"context"
	"fmt"
	"log"
	"strings"

	"pixpivot/arc/config"

	"cloud.google.com/go/translate"
	"golang.org/x/text/language"
	"google.golang.org/api/option"
)

// Supported languages as per 8th Schedule of India + English
var SupportedLanguages = map[string]string{
	"en": "English",
	"as": "Assamese",
	"bn": "Bengali",
	"brx": "Bodo",
	"doi": "Dogri",
	"gu": "Gujarati",
	"hi": "Hindi",
	"kn": "Kannada",
	"ks": "Kashmiri",
	"kok": "Konkani",
	"mai": "Maithili",
	"ml": "Malayalam",
	"mni": "Manipuri",
	"mr": "Marathi",
	"ne": "Nepali",
	"or": "Odia",
	"pa": "Punjabi",
	"sa": "Sanskrit",
	"sat": "Santali",
	"sd": "Sindhi",
	"ta": "Tamil",
	"te": "Telugu",
	"ur": "Urdu",
}

type TranslationService struct {
	client *translate.Client
	apiKey string
}

func NewTranslationService() *TranslationService {
	cfg := config.LoadConfig()
	apiKey := cfg.GoogleTranslateAPIKey

	var client *translate.Client
	var err error

	if apiKey != "" {
		ctx := context.Background()
		client, err = translate.NewClient(ctx, option.WithAPIKey(apiKey))
		if err != nil {
			log.Printf("Failed to create translate client: %v", err)
		}
	} else {
		log.Println("Google Translate API Key not found. Translation service will be disabled.")
	}

	return &TranslationService{
		client: client,
		apiKey: apiKey,
	}
}

func (s *TranslationService) TranslateText(ctx context.Context, text string, targetLang string) (string, error) {
	if s.client == nil {
		return text, fmt.Errorf("translation service not configured")
	}

	// Validate target language
	if _, ok := SupportedLanguages[targetLang]; !ok {
		// Try to match by prefix if full code not found (e.g. "hi-IN" -> "hi")
		parts := strings.Split(targetLang, "-")
		if len(parts) > 0 {
			if _, ok := SupportedLanguages[parts[0]]; ok {
				targetLang = parts[0]
			} else {
				return text, fmt.Errorf("unsupported language: %s", targetLang)
			}
		} else {
			return text, fmt.Errorf("unsupported language: %s", targetLang)
		}
	}

	lang, err := language.Parse(targetLang)
	if err != nil {
		return text, fmt.Errorf("invalid language code: %w", err)
	}

	resp, err := s.client.Translate(ctx, []string{text}, lang, nil)
	if err != nil {
		return text, fmt.Errorf("translation failed: %w", err)
	}

	if len(resp) > 0 {
		return resp[0].Text, nil
	}

	return text, nil
}

func (s *TranslationService) TranslateMap(ctx context.Context, content map[string]string, targetLang string) (map[string]string, error) {
	result := make(map[string]string)
	for key, text := range content {
		translated, err := s.TranslateText(ctx, text, targetLang)
		if err != nil {
			log.Printf("Failed to translate key %s: %v", key, err)
			result[key] = text // Fallback to original
		} else {
			result[key] = translated
		}
	}
	return result, nil
}

func (s *TranslationService) GetSupportedLanguages() map[string]string {
	return SupportedLanguages
}

