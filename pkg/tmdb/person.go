package tmdb

import (
	"context"
	"net/url"
	"strconv"
)

type PersonsPage struct {
	Page         int      `json:"page"`
	Results      []Person `json:"results"`
	TotalPages   int      `json:"total_pages"`
	TotalResults int      `json:"total_results"`
}

type Person struct {
	Adult              bool    `json:"adult"`
	Gender             int     `json:"gender"`
	Id                 int     `json:"id"`
	KnownForDepartment string  `json:"known_for_department"`
	Name               string  `json:"name"`
	OriginalName       string  `json:"original_name"`
	Popularity         float64 `json:"popularity"`
	ProfilePath        *string `json:"profile_path"`
	KnownFor           []struct {
		Adult            bool     `json:"adult"`
		BackdropPath     *string  `json:"backdrop_path"`
		Id               int      `json:"id"`
		Title            string   `json:"title,omitempty"`
		OriginalLanguage string   `json:"original_language"`
		OriginalTitle    string   `json:"original_title,omitempty"`
		Overview         string   `json:"overview"`
		PosterPath       *string  `json:"poster_path"`
		MediaType        string   `json:"media_type"`
		GenreIds         []int    `json:"genre_ids"`
		Popularity       float64  `json:"popularity"`
		ReleaseDate      string   `json:"release_date,omitempty"`
		Video            bool     `json:"video,omitempty"`
		VoteAerage       float64  `json:"vote_aerage,omitempty"`
		VoteCount        int      `json:"vote_count"`
		VoteAverage      float64  `json:"vote_average,omitempty"`
		Name             string   `json:"name,omitempty"`
		OriginalName     string   `json:"original_name,omitempty"`
		FirstAirDate     string   `json:"first_air_date,omitempty"`
		OriginCountry    []string `json:"origin_country,omitempty"`
	} `json:"known_for"`
}

func (c Client) SearchPersonPage(ctx context.Context, query string, page int) ([]Person, int, error) {
	values := url.Values{
		"query": []string{query},
		"page":  []string{strconv.Itoa(page)},
	}

	result, err := call[PersonsPage](ctx, c, c.BaseURL+"/3/search/person", values)
	if err != nil {
		return nil, 0, err
	}
	return result.Results, result.TotalPages, nil
}

func (c Client) SearchPersonAllPages(ctx context.Context, query string) ([]Person, error) {
	var allPersons []Person
	page := 1
	for {
		result, totalPages, err := c.SearchPersonPage(ctx, query, page)
		if err == nil {
			allPersons = append(allPersons, result...)
		}

		if err != nil || page == totalPages {
			return allPersons, err
		}
	}
}

func (c Client) GetPerson(ctx context.Context, id int) (Person, error) {
	return call[Person](ctx, c, c.BaseURL+"/3/person/"+strconv.Itoa(id), url.Values{})
}

type PersonCredits struct {
	Cast []CastCredit `json:"cast"`
	Crew []CrewCredit `json:"crew"`
	Id   int          `json:"id"`
}

func (c Client) GetPersonCredits(ctx context.Context, id int) (PersonCredits, error) {
	return call[PersonCredits](ctx, c, c.BaseURL+"/3/person/"+strconv.Itoa(id)+"/combined_credits", nil)
}

type CastCredit struct {
	Adult            bool     `json:"adult"`
	BackdropPath     *string  `json:"backdrop_path"`
	GenreIds         []int    `json:"genre_ids"`
	Id               int      `json:"id"`
	OriginalLanguage string   `json:"original_language"`
	OriginalTitle    string   `json:"original_title,omitempty"`
	Overview         string   `json:"overview"`
	Popularity       float64  `json:"popularity"`
	PosterPath       *string  `json:"poster_path"`
	ReleaseDate      string   `json:"release_date,omitempty"`
	Title            string   `json:"title,omitempty"`
	Video            bool     `json:"video,omitempty"`
	VoteAverage      float64  `json:"vote_average"`
	VoteCount        int      `json:"vote_count"`
	Character        string   `json:"character"`
	CreditId         string   `json:"credit_id"`
	Order            int      `json:"order,omitempty"`
	MediaType        string   `json:"media_type"`
	OriginCountry    []string `json:"origin_country,omitempty"`
	OriginalName     string   `json:"original_name,omitempty"`
	FirstAirDate     string   `json:"first_air_date,omitempty"`
	Name             string   `json:"name,omitempty"`
	EpisodeCount     int      `json:"episode_count,omitempty"`
}

func (c CastCredit) GetTitle() string {
	switch c.MediaType {
	case "movie":
		return c.Title
	case "tv":
		return c.Name
	default:
		return "unknown"
	}
}

type CrewCredit struct {
	Adult            bool     `json:"adult"`
	BackdropPath     *string  `json:"backdrop_path"`
	GenreIds         []int    `json:"genre_ids"`
	Id               int      `json:"id"`
	OriginalLanguage string   `json:"original_language"`
	OriginalTitle    string   `json:"original_title,omitempty"`
	Overview         string   `json:"overview"`
	Popularity       float64  `json:"popularity"`
	PosterPath       *string  `json:"poster_path"`
	ReleaseDate      string   `json:"release_date,omitempty"`
	Title            string   `json:"title,omitempty"`
	Video            bool     `json:"video,omitempty"`
	VoteAverage      float64  `json:"vote_average"`
	VoteCount        int      `json:"vote_count"`
	CreditId         string   `json:"credit_id"`
	Department       string   `json:"department"`
	Job              string   `json:"job"`
	MediaType        string   `json:"media_type"`
	OriginCountry    []string `json:"origin_country,omitempty"`
	OriginalName     string   `json:"original_name,omitempty"`
	FirstAirDate     string   `json:"first_air_date,omitempty"`
	Name             string   `json:"name,omitempty"`
	EpisodeCount     int      `json:"episode_count,omitempty"`
}
