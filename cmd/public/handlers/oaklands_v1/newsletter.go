package oaklands_v1

import (
	"net/http"
	"slices"
	"sort"
	"strings"

	"github.com/go-chi/chi"
	"github.com/redis/go-redis/v9"
	"github.com/typical-developers/public-experience-api/cmd/public/rest"
	"github.com/typical-developers/public-experience-api/internal/oaklands"
	"github.com/typical-developers/public-experience-api/pkg/httpx"
)

func (o *OaklandsV1Routes) sortNewsletters(values []NewsletterEntry, orderBy string) []NewsletterEntry {
	descending := orderBy == "desc"

	sort.Slice(values, func(i, j int) bool {
		if values[i].Date.Equal(values[j].Date) {
			if descending {
				return values[i].ID > values[j].ID
			}
			return values[i].ID < values[j].ID
		}
		if descending {
			return values[i].Date.After(values[j].Date)
		}

		return values[i].Date.Before(values[j].Date)
	})

	return values
}

//	@Router			/oaklands/v1/newsletters [GET]
//	@Description	Fetch a list of newsletters.
//
//	@Tags			Oaklands
//
//	@Param			If-None-Match	header		string	false	"ETag to validate cached response."
//	@Param			order_by		query		string	false	"The direction to order by. This will use the changelog's date to order."	default(desc)	enums(desc, asc)
//
//	@Success		200				{object}	object{data=[]NewsletterEntry}
//	@Failure		500				{object}	rest.ErrorResponse
//	@Failure		503				{object}	rest.ErrorResponse
//
// swagger:ignore
func (o *OaklandsV1Routes) ListNewsletters(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	orderBy := httpx.QueryGet(r, "order_by", "desc")

	config, err := o.store.Config.Get(ctx)
	if err != nil {
		rest.WriteRESTError(w, err)
		return
	}

	ids, err := o.store.Newsletter.Keys(ctx)
	if err != nil {
		rest.WriteRESTError(w, err)
		return
	}
	if len(ids) == 0 {
		rest.WriteRESTError(w, redis.Nil)
		return
	}

	newsletters, err := o.store.Newsletter.GetMany(ctx, ids...)
	if err != nil {
		rest.WriteRESTError(w, err)
		return
	}

	response := rest.Response[[]NewsletterEntry]{
		Data: make([]NewsletterEntry, 0, len(newsletters)),
	}
	for id, entry := range newsletters {
		if slices.Contains(config.HiddenNewsletters, id) {
			continue
		}
		response.Data = append(response.Data, NewsletterEntry{
			ID:            id,
			Header:        entry.Header,
			BannerImageId: entry.BannerImageId,
			Date:          entry.Date,
		})
	}
	response.Data = o.sortNewsletters(response.Data, orderBy)

	if err := httpx.WriteJSONWithETag(w, r, response, http.StatusOK); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

//	@Router			/oaklands/v1/newsletters/{id} [GET]
//	@Description	Fetch a list of newsletters.
//
//	@Tags			Oaklands
//
//	@Param			If-None-Match	header		string	false	"ETag to validate cached response."
//	@Param			id				path		string	true	"The id of the newsletter. For quick access to the latest newsletter, use 'latest' as the value."	default(latest)
//
//	@Success		200				{object}	object{data=Newsletter}
//	@Failure		500				{object}	rest.ErrorResponse
//	@Failure		503				{object}	rest.ErrorResponse
//
// swagger:ignore
func (o *OaklandsV1Routes) GetNewsletter(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id := chi.URLParam(r, "id")
	if id == "latest" {
		config, err := o.store.Config.Get(ctx)
		if err == nil && config.NewsletterOverride != "" && config.NewsletterOverride != "Default" {
			id = config.NewsletterOverride
		}
	}
	if strings.EqualFold(id, "latest") {
		latest, err := o.store.NewsletterLatest.Get(ctx)
		if err != nil {
			rest.WriteRESTError(w, err)
			return
		}
		id = latest
	}

	newsletter, err := o.store.Newsletter.Get(ctx, id)
	if err != nil {
		rest.WriteRESTError(w, err)
		return
	}

	response := rest.Response[Newsletter]{
		Data: Newsletter{
			Header:        newsletter.Header,
			Subheader:     newsletter.Subheader,
			Date:          newsletter.Date,
			BannerImageId: newsletter.BannerImageId,
			Sections:      make([]NewsletterSection, len(newsletter.Section)),
		},
	}

	for i, section := range newsletter.Section {
		sectionContent := make([]any, len(section.Content))

		for j, content := range section.Content {
			switch v := content.(type) {
			case oaklands.NewsletterParagraph:
				sectionContent[j] = NewsletterParagraph{
					NewsLetterContent: NewsLetterContent{Type: "paragraph"},
					Text:              v.Text,
				}
			case oaklands.NewsletterImage:
				sectionContent[j] = NewsletterImage{
					NewsLetterContent: NewsLetterContent{Type: "image"},
					ImageId:           v.ImageId,
				}
			case oaklands.NewsletterImageCarousel:
				sectionContent[j] = NewsletterImageCarousel{
					NewsLetterContent: NewsLetterContent{Type: "image_carousel"},
					ImageIds:          v.ImageIds,
				}
			case oaklands.NewsletterVideo:
				sectionContent[j] = NewsletterVideo{
					NewsLetterContent: NewsLetterContent{Type: "video"},
					VideoId:           v.VideoId,
				}
			case oaklands.NewsletterBaseContent:
				sectionContent[j] = NewsLetterContent{Type: string(v.Type)}
			}
		}

		response.Data.Sections[i] = NewsletterSection{
			Header:  section.Header,
			Content: sectionContent,
		}
	}

	w.Header().Set("Cache-Control", "public, max-age=0, s-maxage=300, stale-while-revalidate=300")
	if err := httpx.WriteJSONWithETag(w, r, response, http.StatusOK); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}
