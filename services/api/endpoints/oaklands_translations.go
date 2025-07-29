package endpoints

import (
	"sort"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/typical-developers/public-experience-api/internal/cache"
)

//	@Router		/v1/oaklands/translations/keys [GET]
//
//	@Tags		Oaklands, Oaklands - Translations
//
//	@Param		language	query		string	false	"The language to fetch the translation keys for. Only supports 'en_us'."
//	@Param		search		query		string	false	"A search term to filter the translation keys by."
//	@Param		omit		query		string	false	"A omit term to filter the translation keys by."
//
//	@Success	200			{object}	OaklandsTranslationKeys
//
// nolint:staticcheck
func OaklandsV1TranslationKeys(c *fiber.Ctx) error {
	keys := []string{}
	lang := c.Query("language", "en_us")
	search := c.Query("search")
	omit := c.Query("omit")

	var translations *cache.OaklandsTranslations
	if translations = cache.GetCached[cache.OaklandsTranslations](c.Context(), "oaklands:translations", "$"); translations == nil {
		c.Status(fiber.StatusInternalServerError)
		return c.JSON(ErrorAPIResponse{
			Success: false,
			Message: "Unable to fetch translations",
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
		return c.JSON(ErrorAPIResponse{
			Success: false,
			Message: "No translation keys found.",
		})
	}

	sort.Strings(keys)
	return c.JSON(OaklandsTranslationKeys{
		Success: true,
		Data:    keys,
	})
}

//	@Router		/v1/oaklands/translations [GET]
//
//	@Tags		Oaklands, Oaklands - Translations
//
//	@Param		language	query		string		false	"The language to translate the strings into. Only supports 'en_us'."
//	@Param		strings		query		[]string	true	"A comma-separated list of strings return translations for."
//
//	@Success	200			{object}	OaklandsTranslations
//
// nolint:staticcheck
func OaklandsV1Translations(c *fiber.Ctx) error {
	strs := c.Query("strings")
	lang := c.Query("language", "en_us")

	keys := strings.Split(strs, ",")

	if len(keys) <= 1 && keys[0] == "" {
		c.Status(fiber.StatusBadRequest)
		return c.JSON(ErrorAPIResponse{
			Success: false,
			Message: "No strings provided.",
		})
	}

	var translations *cache.OaklandsTranslations
	if translations = cache.GetCached[cache.OaklandsTranslations](c.Context(), "oaklands:translations", "$"); translations == nil {
		c.Status(fiber.StatusServiceUnavailable)
		return c.JSON(ErrorAPIResponse{
			Success: false,
			Message: "Unable to fetch translations.",
		})
	}

	langTranslations := (*translations)[lang]
	filteredKeys := make(map[string]string)
	for _, key := range keys {
		if _, ok := langTranslations[key]; !ok {
			continue
		}

		filteredKeys[key] = langTranslations[key].(string)
	}

	return c.JSON(OaklandsTranslations{
		Success: true,
		Data:    filteredKeys,
	})
}
