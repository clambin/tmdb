package tmdb

import (
	"context"
	"net/url"
	"strconv"
)

type TVSeries struct {
	Adult        bool   `json:"adult"`
	BackdropPath string `json:"backdrop_path"`
	CreatedBy    []struct {
		Id           int    `json:"id"`
		CreditId     string `json:"credit_id"`
		Name         string `json:"name"`
		OriginalName string `json:"original_name"`
		Gender       int    `json:"gender"`
		ProfilePath  string `json:"profile_path"`
	} `json:"created_by"`
	EpisodeRunTime []interface{} `json:"episode_run_time"`
	FirstAirDate   string        `json:"first_air_date"`
	Genres         []struct {
		Id   int    `json:"id"`
		Name string `json:"name"`
	} `json:"genres"`
	Homepage         string   `json:"homepage"`
	Id               int      `json:"id"`
	InProduction     bool     `json:"in_production"`
	Languages        []string `json:"languages"`
	LastAirDate      string   `json:"last_air_date"`
	LastEpisodeToAir struct {
		Id             int     `json:"id"`
		Overview       string  `json:"overview"`
		Name           string  `json:"name"`
		VoteAverage    float64 `json:"vote_average"`
		VoteCount      int     `json:"vote_count"`
		AirDate        string  `json:"air_date"`
		EpisodeNumber  int     `json:"episode_number"`
		EpisodeType    string  `json:"episode_type"`
		ProductionCode string  `json:"production_code"`
		Runtime        int     `json:"runtime"`
		SeasonNumber   int     `json:"season_number"`
		ShowId         int     `json:"show_id"`
		StillPath      string  `json:"still_path"`
	} `json:"last_episode_to_air"`
	Name             string      `json:"name"`
	NextEpisodeToAir interface{} `json:"next_episode_to_air"`
	Networks         []struct {
		Id            int    `json:"id"`
		LogoPath      string `json:"logo_path"`
		Name          string `json:"name"`
		OriginCountry string `json:"origin_country"`
	} `json:"networks"`
	NumberOfEpisodes    int      `json:"number_of_episodes"`
	NumberOfSeasons     int      `json:"number_of_seasons"`
	OriginCountry       []string `json:"origin_country"`
	OriginalLanguage    string   `json:"original_language"`
	OriginalName        string   `json:"original_name"`
	Overview            string   `json:"overview"`
	Popularity          float64  `json:"popularity"`
	PosterPath          string   `json:"poster_path"`
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
	Seasons []struct {
		AirDate      string  `json:"air_date"`
		EpisodeCount int     `json:"episode_count"`
		Id           int     `json:"id"`
		Name         string  `json:"name"`
		Overview     string  `json:"overview"`
		PosterPath   string  `json:"poster_path"`
		SeasonNumber int     `json:"season_number"`
		VoteAverage  float64 `json:"vote_average"`
	} `json:"seasons"`
	SpokenLanguages []struct {
		EnglishName string `json:"english_name"`
		Iso6391     string `json:"iso_639_1"`
		Name        string `json:"name"`
	} `json:"spoken_languages"`
	Status      string  `json:"status"`
	Tagline     string  `json:"tagline"`
	Type        string  `json:"type"`
	VoteAverage float64 `json:"vote_average"`
	VoteCount   int     `json:"vote_count"`
}

func (c Client) GetTVSeries(ctx context.Context, id int) (TVSeries, error) {
	return call[TVSeries](ctx, c, c.BaseURL+"/3/tv/"+strconv.Itoa(id), url.Values{})
}

type TVSeriesCredits struct {
	Cast []TVSeriesCastCredits `json:"cast"`
	Crew []TVSeriesCrewCredits `json:"crew"`
	Id   int                   `json:"id"`
}

type TVSeriesCastCredits struct {
	Adult              bool    `json:"adult"`
	Gender             int     `json:"gender"`
	Id                 int     `json:"id"`
	KnownForDepartment string  `json:"known_for_department"`
	Name               string  `json:"name"`
	OriginalName       string  `json:"original_name"`
	Popularity         float64 `json:"popularity"`
	ProfilePath        *string `json:"profile_path"`
	Roles              []struct {
		CreditId     string `json:"credit_id"`
		Character    string `json:"character"`
		EpisodeCount int    `json:"episode_count"`
	} `json:"roles"`
	TotalEpisodeCount int `json:"total_episode_count"`
	Order             int `json:"order"`
}

type TVSeriesCrewCredits struct {
	Adult              bool    `json:"adult"`
	Gender             int     `json:"gender"`
	Id                 int     `json:"id"`
	KnownForDepartment string  `json:"known_for_department"`
	Name               string  `json:"name"`
	OriginalName       string  `json:"original_name"`
	Popularity         float64 `json:"popularity"`
	ProfilePath        *string `json:"profile_path"`
	Jobs               []struct {
		CreditId     string `json:"credit_id"`
		Job          string `json:"job"`
		EpisodeCount int    `json:"episode_count"`
	} `json:"jobs"`
	Department        string `json:"department"`
	TotalEpisodeCount int    `json:"total_episode_count"`
}

func (c Client) GetTVSeriesCredits(ctx context.Context, id int) (TVSeriesCredits, error) {
	return call[TVSeriesCredits](ctx, c, c.BaseURL+"/3/tv/"+strconv.Itoa(id)+"/aggregate_credits", url.Values{})
}
