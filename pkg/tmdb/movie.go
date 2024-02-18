package tmdb

import (
	"context"
	"net/url"
	"strconv"
)

type Movie struct {
	Adult               bool        `json:"adult"`
	BackdropPath        interface{} `json:"backdrop_path"`
	BelongsToCollection struct {
		Id           int    `json:"id"`
		Name         string `json:"name"`
		PosterPath   string `json:"poster_path"`
		BackdropPath string `json:"backdrop_path"`
	} `json:"belongs_to_collection"`
	Budget int `json:"budget"`
	Genres []struct {
		Id   int    `json:"id"`
		Name string `json:"name"`
	} `json:"genres"`
	Homepage            string      `json:"homepage"`
	Id                  int         `json:"id"`
	ImdbId              interface{} `json:"imdb_id"`
	OriginalLanguage    string      `json:"original_language"`
	OriginalTitle       string      `json:"original_title"`
	Overview            string      `json:"overview"`
	Popularity          float64     `json:"popularity"`
	PosterPath          string      `json:"poster_path"`
	ProductionCompanies []struct {
		Id            int    `json:"id"`
		LogoPath      string `json:"logo_path"`
		Name          string `json:"name"`
		OriginCountry string `json:"origin_country"`
	} `json:"production_companies"`
	ProductionCountries []struct {
		Iso31661 string `json:"iso_3166_1"`
		Name     string `json:"name"`
	} `json:"production_countries"`
	ReleaseDate     string `json:"release_date"`
	Revenue         int    `json:"revenue"`
	Runtime         int    `json:"runtime"`
	SpokenLanguages []struct {
		EnglishName string `json:"english_name"`
		Iso6391     string `json:"iso_639_1"`
		Name        string `json:"name"`
	} `json:"spoken_languages"`
	Status      string  `json:"status"`
	Tagline     string  `json:"tagline"`
	Title       string  `json:"title"`
	Video       bool    `json:"video"`
	VoteAverage float64 `json:"vote_average"`
	VoteCount   int     `json:"vote_count"`
}

func (c Client) GetMovie(ctx context.Context, id int) (Movie, error) {
	return call[Movie](ctx, c, c.BaseURL+"/3/movie/"+strconv.Itoa(id), url.Values{})
}

type MovieCredits struct {
	Id   int                `json:"id"`
	Cast []MovieCastCredits `json:"cast"`
	Crew []MovieCrewCredits `json:"crew"`
}

func (c Client) GetMovieCredits(ctx context.Context, id int) (MovieCredits, error) {
	return call[MovieCredits](ctx, c, c.BaseURL+"/3/movie/"+strconv.Itoa(id)+"/credits", url.Values{})
}

type MovieCastCredits struct {
	Adult              bool    `json:"adult"`
	Gender             int     `json:"gender"`
	Id                 int     `json:"id"`
	KnownForDepartment string  `json:"known_for_department"`
	Name               string  `json:"name"`
	OriginalName       string  `json:"original_name"`
	Popularity         float64 `json:"popularity"`
	ProfilePath        *string `json:"profile_path"`
	CastId             int     `json:"cast_id"`
	Character          string  `json:"character"`
	CreditId           string  `json:"credit_id"`
	Order              int     `json:"order"`
}

type MovieCrewCredits struct {
	Adult              bool    `json:"adult"`
	Gender             int     `json:"gender"`
	Id                 int     `json:"id"`
	KnownForDepartment string  `json:"known_for_department"`
	Name               string  `json:"name"`
	OriginalName       string  `json:"original_name"`
	Popularity         float64 `json:"popularity"`
	ProfilePath        *string `json:"profile_path"`
	CreditId           string  `json:"credit_id"`
	Department         string  `json:"department"`
	Job                string  `json:"job"`
}
