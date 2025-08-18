package utility

import (
	"strings"

	"github.com/gofiber/fiber/v2/log"
	"github.com/pemistahl/lingua-go"
)

func LangDetect(text string) lingua.Language {
	query := strings.TrimSpace(text)

	languages := []lingua.Language{lingua.English, lingua.Russian}
	detector := lingua.NewLanguageDetectorBuilder().
		FromLanguages(languages...).
		Build()

	lang, err := detector.DetectLanguageOf(query)
	if err == false {
		log.Error("Error while detecting language:", err)
		return lingua.Russian
	}
	return lang
}
