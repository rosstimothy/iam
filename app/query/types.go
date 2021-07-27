package query

type Role struct {
	Name        string   `json:"name"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
	Stage       int      `json:"stage"`
	Etag        string   `json:"etag"`
}
