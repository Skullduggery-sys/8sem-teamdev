package service

type movieResponse struct {
	Name string `json:"name"`
	Year int    `json:"year"`
	// Rating       ratingResponse  `json:"rating"`
	Poster       imageResponse   `json:"poster"`
	Genres       []genreResponse `json:"genres"`
	IsSeries     bool            `json:"isSeries"`
	MovieLength  int             `json:"movieLength"`
	SeriesLength int             `json:"seriesLength"`
	SeasonsInfo  []seasonInfo    `json:"seasonsInfo"`
}

// type ratingResponse struct {
// 	KP float64 `json:"kp"`
// }

type imageResponse struct {
	URL string `json:"url"`
}

type genreResponse struct {
	Name string `json:"name"`
}

type seasonInfo struct {
	Number        int `json:"number"`
	EpisodesCount int `json:"episodesCount"`
}
