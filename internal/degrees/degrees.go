package degrees

import (
	"context"
	"github.com/clambin/tmdb/pkg/tmdb"
	"log/slog"
	"sync"
	"sync/atomic"
)

type TMDBClient interface {
	SearchPersonPage(ctx context.Context, query string, page int) ([]tmdb.Person, int, error)
	SearchPersonAllPages(ctx context.Context, query string) ([]tmdb.Person, error)
	GetPerson(ctx context.Context, id int) (tmdb.Person, error)
	GetPersonCredits(ctx context.Context, id int) (tmdb.PersonCredits, error)
	GetMovieCredits(ctx context.Context, id int) (tmdb.MovieCredits, error)
	GetTVSeriesCredits(ctx context.Context, id int) (tmdb.TVSeriesCredits, error)
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////

type Mode int

const (
	ModeMovies   Mode = 0x1
	ModeTVSeries Mode = 0x2
)

type PathFinder struct {
	TMDBClient TMDBClient
	From       tmdb.Person
	To         tmdb.Person
	Logger     *slog.Logger
	Mode       Mode

	maxPathLength          atomic.Int32
	toActorMovieCredits    actorCredits
	toActorTVSeriesCredits actorCredits
	lock                   sync.RWMutex
	visitedActors          map[int]struct{}
}

func (f *PathFinder) Find(ctx context.Context, ch chan Path, depth int) {
	defer close(ch)
	if err := f.init(ctx, depth+1); err != nil {
		f.Logger.Error("failed to initialize path finder", "error", err)
		return
	}
	f.findActor(ctx, ch, f.From, Path{Link{Person: f.From}})
}

func (f *PathFinder) init(ctx context.Context, maxPathLength int) (err error) {
	f.maxPathLength.Store(int32(maxPathLength))
	f.visitedActors = make(map[int]struct{})
	f.toActorMovieCredits, f.toActorTVSeriesCredits, err = f.getActorCredits(ctx, f.To.Id)
	return err
}

func (f *PathFinder) findActor(ctx context.Context, ch chan Path, from tmdb.Person, path Path) {
	logger := f.Logger.With("path", path)
	logger.Debug("checking actor", "actor", from.Name)

	fromActorMovieCredits, fromActorTVSeriesCredits, err := f.getActorCredits(ctx, from.Id)
	if err != nil {
		logger.Error("failed to get actor credits", "err", err)
		return
	}

	f.visit(from.Id)

	var found bool
	for id, title := range commonActorCredits(fromActorMovieCredits, f.toActorMovieCredits) {
		f.report(ch, append(path, Link{Movie: Movie{ID: id, Name: title}, Person: f.To}))
		found = true
	}
	for id, title := range commonActorCredits(fromActorTVSeriesCredits, f.toActorTVSeriesCredits) {
		f.report(ch, append(path, Link{Movie: Movie{ID: id, Name: title}, Person: f.To}))
		found = true
	}

	if found || f.maxPathReached(path) {
		return
	}

	var wg sync.WaitGroup
	if f.Mode&ModeMovies != 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			f.findActorInMovies(ctx, ch, path, fromActorMovieCredits, logger)
		}()
	}
	if f.Mode&ModeTVSeries != 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			f.findActorInTVSeries(ctx, ch, path, fromActorTVSeriesCredits, logger)
		}()
	}
	wg.Wait()
}

func (f *PathFinder) findActorInMovies(ctx context.Context, ch chan Path, path Path, actorCredits actorCredits, logger *slog.Logger) {
	var wg sync.WaitGroup
	for id, title := range actorCredits {
		credits, err := f.TMDBClient.GetMovieCredits(ctx, id)
		if err != nil {
			logger.Error("failed to get movie credits", "id", id, "err", err)
			return
		}
		for _, cast := range credits.Cast {
			// other goroutines may find a shorter path concurrently
			if f.maxPathReached(path) {
				break
			}
			// don't traverse actors we've already visited
			if f.visited(cast.Id) {
				continue
			}
			p := tmdb.Person{Id: cast.Id, Name: cast.Name}
			newPath := make(Path, len(path)+1)
			copy(newPath, append(path, Link{
				Movie:  Movie{ID: id, Name: title},
				Person: p,
			}))
			wg.Add(1)
			go func() {
				f.findActor(ctx, ch, p, newPath)
				wg.Done()
			}()
		}
	}
	wg.Wait()
}

func (f *PathFinder) findActorInTVSeries(ctx context.Context, ch chan Path, path Path, fromActorTVSeriesCredits actorCredits, logger *slog.Logger) {
	var wg sync.WaitGroup
	for id, title := range fromActorTVSeriesCredits {
		credits, err := f.TMDBClient.GetTVSeriesCredits(ctx, id)
		if err != nil {
			logger.Error("failed to get tv show credits", "id", id, "err", err)
			return
		}
		for _, cast := range credits.Cast {
			// other goroutines may find a shorter path concurrently
			if f.maxPathReached(path) {
				break
			}
			// don't traverse actors we've already visited
			if f.visited(cast.Id) {
				continue
			}
			p := tmdb.Person{Id: cast.Id, Name: cast.Name}
			newPath := make(Path, len(path)+1)
			copy(newPath, append(path, Link{
				Movie:  Movie{ID: id, Name: title},
				Person: p,
			}))
			wg.Add(1)
			go func() {
				f.findActor(ctx, ch, p, newPath)
				wg.Done()
			}()
		}
	}
	wg.Wait()
}

func (f *PathFinder) report(ch chan Path, path Path) {
	if pathLen := len(path); pathLen < int(f.maxPathLength.Load()) {
		f.Logger.Info("shorter path found!", "maxPathLength", pathLen)
		f.maxPathLength.Store(int32(pathLen))
	}
	ch <- path
}

func (f *PathFinder) maxPathReached(path Path) bool {
	return len(path)+1 >= int(f.maxPathLength.Load())
}

func (f *PathFinder) getActorCredits(ctx context.Context, id int) (movies actorCredits, tvSeries actorCredits, err error) {
	movies = make(actorCredits)
	tvSeries = make(actorCredits)
	var credits tmdb.PersonCredits
	if credits, err = f.TMDBClient.GetPersonCredits(ctx, id); err == nil {
		for _, credit := range credits.Cast {
			if !ignoreGenre(credit) {
				if credit.MediaType == "movie" {
					movies.add(credit.Id, credit.GetTitle())
				} else {
					tvSeries.add(credit.Id, credit.GetTitle())
				}
			}
		}
	}
	return movies, tvSeries, err
}

func (f *PathFinder) visit(id int) {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.visitedActors[id] = struct{}{}
}

func (f *PathFinder) visited(id int) bool {
	f.lock.RLock()
	defer f.lock.RUnlock()
	_, ok := f.visitedActors[id]
	return ok
}

var skipList = map[int]struct{}{
	10767: {},
	10763: {},
	99:    {},
}

func ignoreGenre(castCredit tmdb.CastCredit) bool {
	for _, genre := range castCredit.GenreIds {
		if _, ok := skipList[genre]; ok {
			return true
		}
	}
	return false
}
