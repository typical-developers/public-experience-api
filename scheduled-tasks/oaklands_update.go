package tasks

import (
	"context"

	"maps"

	"github.com/typical-developers/public-experience-api/internal/cache"
	"github.com/typical-developers/public-experience-api/internal/luau"
)

func CheckOaklandsUpdates() {
	ctx := context.Background()

	fetchOaklandsTranslations(ctx)
}

func fetchOaklandsTranslations(ctx context.Context) {
	res, err := luau.Run[map[string]any](ctx, "3666294218", "9938675423", luau.OaklandsTranslationsScript)
	if err != nil {
		return
	}

	translations := make(map[string]map[string]any)
	for lang, keys := range *res {
		if lang == "ALTKEYS" {
			continue
		}

		translations[lang] = make(map[string]any)
		maps.Copy(translations[lang], keys.(map[string]any))
	}

	cache.SetCached(ctx, "oaklands:translations", "$", translations, nil)
}
