package rely

type RetJson struct {
	Code         int    `json:"code"`
	ID           string `json:"id"`
	Imgid        string `json:"imgid"`
	RelativePath string `json:"relative_path"`
	URL          string `json:"url"`
	ThumbnailURL string `json:"thumbnail_url"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	Delete       string `json:"delete"`
}
