package oaklands_v1

import (
	"errors"
	"net/http"
	"sort"
	"time"

	"github.com/go-chi/chi"
	"github.com/redis/go-redis/v9"
	models "github.com/typical-developers/public-experience-api/cmd/public/handlers"
	"github.com/typical-developers/public-experience-api/internal/oaklands"
	"github.com/typical-developers/public-experience-api/pkg/httpx"
)

func (o *OaklandsV1Routes) sortNewsletters(values []oaklands.Newsletters, orderBy string) []oaklands.Newsletters {
	descending := orderBy == "desc"

	sort.Slice(values, func(i, j int) bool {
		if descending {
			return values[i].Date.After(values[j].Date)
		}

		return values[i].Date.Before(values[j].Date)
	})

	return values
}

//	@Router			/v1/oaklands/newsletters [GET]
//	@Description	Fetch a list of newsletters.
//
//	@Tags			Oaklands
//
//	@Param			If-None-Match	header		string	false	"ETag to validate cached response."
//	@Param			order_by		query		string	false	"The direction to order by. This will use the changelog's date to order."	default(desc)	enums(desc, asc)
//
//	@Success		200				{object}	object{data=[]NewsletterEntry}
//	@Failure		503				{object}	models.ErrorResponse
//	@Failure		500				{object}	models.ErrorResponse
//
// swagger:ignore
func (o *OaklandsV1Routes) GetNewsletters(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	orderBy := httpx.QueryGet(r, "order_by", "desc")

	newsletters, err := o.uc.GetNewsletters(ctx)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			_ = httpx.WriteJSON(w, models.ErrorResponse{
				Type:    "ResourceNotCached",
				Message: "The requested resource is not cached. Try again in a bit.",
			}, http.StatusServiceUnavailable)

			return
		}

		_ = httpx.WriteJSON(w, models.ErrorResponse{
			Type:    "InternalServerError",
			Message: "There was an internal server error, try again later.",
		}, http.StatusInternalServerError)

		return
	}

	newsletters = o.sortNewsletters(newsletters, orderBy)
	response := models.Response[[]NewsletterEntry]{
		Data: make([]NewsletterEntry, len(newsletters)),
	}

	for i, entry := range newsletters {
		response.Data[i] = NewsletterEntry{
			ID:   entry.ID,
			Date: entry.Date,
		}
	}

	if err := httpx.WriteJSONWithETag(w, r, response, http.StatusOK); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

//	@Router			/v1/oaklands/newsletters/{id} [GET]
//	@Description	Fetch a list of newsletters.
//
//	@Tags			Oaklands
//
//	@Param			If-None-Match	header		string	false	"ETag to validate cached response."
//
//	@Success		200				{object}	object{data=Newsletter}
//	@Failure		503				{object}	models.ErrorResponse
//	@Failure		500				{object}	models.ErrorResponse
//
// swagger:ignore
func (o *OaklandsV1Routes) GetNewsletter(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id := chi.URLParam(r, "id")
	newsletter, err := o.uc.GetNewsletter(ctx, id)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			_ = httpx.WriteJSON(w, models.ErrorResponse{
				Type:    "NotFound",
				Message: "This newsletter does not exist.",
			}, http.StatusNotFound)

			return
		}

		_ = httpx.WriteJSON(w, models.ErrorResponse{
			Type:    "InternalServerError",
			Message: "There was an internal server error, try again later.",
		}, http.StatusInternalServerError)

		return
	}

	releaseDate, _ := time.Parse(time.RFC3339, newsletter.DateToISO8601())
	response := models.Response[Newsletter]{
		Data: Newsletter{
			Header:        newsletter.Header,
			Subheader:     newsletter.Subheader,
			Date:          releaseDate,
			BannerImageId: newsletter.BannerImageId,
			Sections:      make([]NewsletterSection, len(newsletter.Sections)),
		},
	}

	for i, section := range newsletter.Sections {
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

	if err := httpx.WriteJSONWithETag(w, r, response, http.StatusOK); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}
