package api

import (
	"sort"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/typical-developers/public-experience-api/internal/cache"
)

func OaklandsTranslationKeysV1(c *fiber.Ctx) error {
	keys := []string{}
	lang := c.Query("language", "en_us")
	search := c.Query("search")
	omit := c.Query("omit")

	var translations *cache.Translations
	if translations = cache.GetCached[cache.Translations](c.Context(), "oaklands:translations", "$"); translations == nil {
		c.Status(fiber.StatusServiceUnavailable)
		return c.JSON(APIResponse[any]{
			Success: false,
			Message: Ptr("Unable to fetch translations."),
		})
	}

	for k := range *translations {
		if k != lang {
			continue
		}

		for key := range (*translations)[k] {
			if search != "" && !strings.Contains(key, search) {
				continue
			}

			if omit != "" && strings.Contains(key, omit) {
				continue
			}

			keys = append(keys, key)
		}

		break
	}

	if len(keys) <= 0 {
		c.Status(fiber.StatusNotFound)
		return c.JSON(APIResponse[any]{
			Success: false,
			Message: Ptr("No translation keys found."),
		})
	}

	sort.Strings(keys)
	return c.JSON(APIResponse[[]string]{
		Success: true,
		Data:    Ptr(keys),
	})
}

func OaklandsTranslationsV1(c *fiber.Ctx) error {
	strs := c.Query("strings")
	lang := c.Query("language", "en_us")

	keys := strings.Split(strs, ",")

	if len(keys) <= 1 && keys[0] == "" {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(APIResponse[any]{
			Success: false,
			Message: Ptr("No strings provided."),
		})
	}

	var translations *cache.Translations
	if translations = cache.GetCached[cache.Translations](c.Context(), "oaklands:translations", "$"); translations == nil {
		c.Status(fiber.StatusServiceUnavailable)
		return c.JSON(APIResponse[any]{
			Success: false,
			Message: Ptr("Unable to fetch translations."),
		})
	}

	langTranslations := (*translations)[lang]
	filteredKeys := make(map[string]any)
	for _, key := range keys {
		if _, ok := langTranslations[key]; !ok {
			continue
		}

		filteredKeys[key] = langTranslations[key]
	}

	return c.JSON(APIResponse[map[string]any]{
		Success: true,
		Data:    Ptr(filteredKeys),
	})
}
