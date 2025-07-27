package jobs

type OaklandsUpdateBinaryInput struct {
	CachedNewsletters map[string]bool `json:"CachedNewsletters"`
	CachedChangelogs  map[string]bool `json:"CachedChangelogs"`
}
