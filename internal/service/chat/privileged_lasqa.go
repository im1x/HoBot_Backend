package chat

import (
	DB "HoBot_Backend/internal/mongo"
	"HoBot_Backend/internal/utility"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/adrg/strutil"
	"github.com/adrg/strutil/metrics"
	"github.com/gofiber/fiber/v2/log"
	"github.com/pemistahl/lingua-go"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type MovieKp struct {
	Id      int       `bson:"_id"`
	TitleEn string    `bson:"title_en"`
	TitleRu string    `bson:"title_ru"`
	Rating  int       `bson:"rating"`
	Date    time.Time `bson:"date"`
}

type MovieRating struct {
	movie MovieKp
	rank  float64
}

type MoviesCache struct {
	mu         sync.RWMutex
	lastUpdate time.Time
	movies     []MovieKp
}

type DaStatus struct {
	isOnline      bool
	isInitialized bool
}

var (
	moviesCache MoviesCache
	cacheTTL    = 1 * time.Hour
	daStatus    DaStatus
)

func (c *MoviesCache) refreshCache() ([]MovieKp, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if time.Since(c.lastUpdate) < cacheTTL {
		return c.movies, nil
	}

	newData, err := getMoviesFromDb()
	if err != nil {
		log.Error("Error while getting movies from db:", err)
		if c.movies != nil {
			return c.movies, nil
		}
		return nil, err
	}

	c.movies = newData
	c.lastUpdate = time.Now()
	return c.movies, nil
}

func (c *MoviesCache) getMovies() ([]MovieKp, error) {
	c.mu.RLock()

	if time.Since(c.lastUpdate) < cacheTTL {
		data := c.movies
		c.mu.RUnlock()
		log.Info("Got movies from cache")
		return data, nil
	}
	c.mu.RUnlock()
	log.Info("Got movies from db")
	return c.refreshCache()
}

func (ds *DaStatus) SetStatus(online bool) (statusChanged bool) {
	statusChanged = ds.isInitialized && online && !ds.isOnline
	ds.isOnline = online
	ds.isInitialized = true
	return
}

func getMoviesFromDb() ([]MovieKp, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := DB.GetCollection(DB.PrivilegedLasqaKp).Find(ctx, bson.M{})
	if err != nil {
		log.Error("Error while finding movies:", err)
		return nil, err
	}

	var movies []MovieKp
	if err = cursor.All(ctx, &movies); err != nil {
		log.Error("Error while decoding movies:", err)
		return nil, err
	}
	return movies, nil
}

func getMovies() ([]MovieKp, error) {
	return moviesCache.getMovies()
}

func searchMovies(movieList []MovieKp, query string, lang lingua.Language) []MovieRating {
	var searchResults []MovieRating

	swg := metrics.NewSmithWatermanGotoh()
	swg.CaseSensitive = false
	swg.GapPenalty = -0.1
	swg.Substitution = metrics.MatchMismatch{
		Match:    1,
		Mismatch: -0.5,
	}

	highSimilarity := 0.0

	for _, mv := range movieList {
		if similarity := strutil.Similarity(query, getTitle(mv, lang), swg); similarity >= 0.6 {
			if similarity > highSimilarity {
				highSimilarity = similarity
			}
			searchResults = append(searchResults, MovieRating{mv, similarity})
		}
	}

	if highSimilarity == 1.0 {
		var temp []MovieRating
		for _, m := range searchResults {
			if m.rank == highSimilarity {
				temp = append(temp, m)
			}
		}
		searchResults = temp
	} else {
		sort.Slice(searchResults, func(i, j int) bool {
			return searchResults[i].rank > searchResults[j].rank
		})

		searchResults = slices.Collect(func(yield func(MovieRating) bool) {
			for _, m := range searchResults {
				if highSimilarity >= m.rank+0.11 {
					return
				}
				if !yield(m) {
					return
				}
			}
		})

	}

	return searchResults[:min(5, len(searchResults))]
}

func getTitle(movie MovieKp, lang lingua.Language) string {
	if lang == lingua.Russian {
		return movie.TitleRu
	}
	return movie.TitleEn
}

func formatMsg(m []MovieRating, lang lingua.Language) string {
	numEmoji := map[int]string{
		0: "1️⃣",
		1: "2️⃣",
		2: "3️⃣",
		3: "4️⃣",
		4: "5️⃣",
	}
	result := ""
	resFormat := "%s&ensp;🌟%d&ensp;📅%s"
	dataFormat := "02.01.2006 15:04"
	if len(m) == 1 {
		result += fmt.Sprintf(resFormat,
			getTitle(m[0].movie, lang),
			m[0].movie.Rating,
			m[0].movie.Date.Format(dataFormat))
	} else {
		for i, mv := range m {
			result += fmt.Sprintf("%s "+resFormat+" &#12288;&#12288;",
				numEmoji[i],
				getTitle(mv.movie, lang),
				mv.movie.Rating,
				mv.movie.Date.Format(dataFormat))
		}
	}

	return result
}

func lasqaKp(msg *ChatMsg, param string) {
	_, rest := getAliasAndRestFromMessage(param)
	if rest == "" {
		SendWhisperToUser("🎬🍿 Добавьте стримера в друзья на Кинопоиске (https://www.kinopoisk.ru/user/1059598) "+
			"чтобы видеть, какие фильмы он уже смотрел, или используйте команду !кп <название фильма>", msg.GetChannelId(), msg.GetUser())
		return
	}

	movies, err := getMovies()
	if err != nil {
		log.Error("Error while getting movies:", err)
		return
	}

	if rest == "инфо" {
		sort.Slice(movies, func(i, j int) bool {
			return movies[i].Date.After(movies[j].Date)
		})
		SendWhisperToUser("🎬🍿 Всего фильмов - "+strconv.Itoa(len(movies))+
			" Последний добавленный фильм: "+movies[0].TitleRu, msg.GetChannelId(), msg.GetUser())
		return
	}

	lang := utility.LangDetect(rest)
	sMov := searchMovies(movies, rest, lang)

	if len(sMov) == 0 {
		SendWhisperToUser("🎬🍿 Ничего не нашлось", msg.GetChannelId(), msg.GetUser())
		return
	}

	SendWhisperToUser(formatMsg(sMov, lang), msg.GetChannelId(), msg.GetUser())
}

func CheckDonationAlertsStatus() {
	resp, err := http.Get("https://www.donationalerts.com/api/v1/user/lasqa/donationpagesettings")
	if err != nil || resp.StatusCode != 200 {
		log.Error("Error while check DA status")
		return
	}
	defer resp.Body.Close()

	type DaResponse struct {
		Data struct {
			IsOnline int `json:"is_online"`
		} `json:"data"`
	}

	var daResp DaResponse
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("Error while check DA status, read body:", err)
		return
	}

	err = json.Unmarshal(b, &daResp)
	if err != nil {
		log.Error("Error while unmarshal DA response:", err)
	}

	isOnline := daResp.Data.IsOnline == 1

	if changed := daStatus.SetStatus(isOnline); changed {
		SendMessageToChannel("👀 Стример зашел в DonationAlerts", "8845069", nil)
	}
}
