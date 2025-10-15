package chat

import (
	"HoBot_Backend/internal/model"
	"HoBot_Backend/internal/utility"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"sort"
	"strconv"
	"sync"
	"time"

	repoPrivilegedLasqaKp "HoBot_Backend/internal/repository/privilegedlasqakp"
	repoUser "HoBot_Backend/internal/repository/user"

	"github.com/adrg/strutil"
	"github.com/adrg/strutil/metrics"
	"github.com/gofiber/fiber/v2/log"
	"github.com/pemistahl/lingua-go"
)

type MovieRating struct {
	movie model.MovieKp
	rank  float64
}

type MoviesCache struct {
	mu         sync.RWMutex
	lastUpdate time.Time
	movies     []model.MovieKp
}

type DaStatus struct {
	isOnline      bool
	isInitialized bool
}

type LasqaService struct {
	moviesCache MoviesCache
	cacheTTL    time.Duration
	daStatus    DaStatus
	lasqaKpRepo repoPrivilegedLasqaKp.Repository
	userRepo    repoUser.Repository
	chatService *ChatService
}

/* var (
	moviesCache MoviesCache
	cacheTTL    = 1 * time.Hour
	daStatus    DaStatus
) */

func NewLasqaService(lasqaKpRepo repoPrivilegedLasqaKp.Repository, userRepo repoUser.Repository, chatService *ChatService) *LasqaService {
	return &LasqaService{
		moviesCache: MoviesCache{},
		cacheTTL:    1 * time.Hour,
		daStatus:    DaStatus{},
		lasqaKpRepo: lasqaKpRepo,
		userRepo:    userRepo,
		chatService: chatService,
	}
}

func (s *LasqaService) refreshCache() ([]model.MovieKp, error) {
	s.moviesCache.mu.Lock()
	defer s.moviesCache.mu.Unlock()

	if time.Since(s.moviesCache.lastUpdate) < s.cacheTTL {
		return s.moviesCache.movies, nil
	}

	newData, err := s.lasqaKpRepo.GetMovies()
	if err != nil {
		log.Error("Error while getting movies from db:", err)
		if s.moviesCache.movies != nil {
			return s.moviesCache.movies, nil
		}
		return nil, err
	}

	s.moviesCache.movies = newData
	s.moviesCache.lastUpdate = time.Now()
	return s.moviesCache.movies, nil
}

func (s *LasqaService) getMovies() ([]model.MovieKp, error) {
	s.moviesCache.mu.RLock()

	if time.Since(s.moviesCache.lastUpdate) < s.cacheTTL {
		data := s.moviesCache.movies
		s.moviesCache.mu.RUnlock()
		return data, nil
	}
	s.moviesCache.mu.RUnlock()
	return s.refreshCache()
}

func (s *LasqaService) SetStatus(online bool) (statusChanged bool) {
	statusChanged = s.daStatus.isInitialized && online && !s.daStatus.isOnline
	s.daStatus.isOnline = online
	s.daStatus.isInitialized = true
	return
}

func (s *LasqaService) searchMovies(movieList []model.MovieKp, query string, lang lingua.Language) []MovieRating {
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

func getTitle(movie model.MovieKp, lang lingua.Language) string {
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

func (s *LasqaService) LasqaKp(msg *model.ChatMsg, param string) {
	_, rest := getAliasAndRestFromMessage(param)
	if rest == "" {
		s.chatService.SendWhisperToUser("🎬🍿 Добавьте стримера в друзья на Кинопоиске (https://www.kinopoisk.ru/user/1059598) "+
			"чтобы видеть, какие фильмы он уже смотрел, или используйте команду !кп <название фильма>", msg.GetChannelId(s.userRepo.GetUserIdByWs), msg.GetUser())
		return
	}

	movies, err := s.getMovies()
	if err != nil {
		log.Error("Error while getting movies:", err)
		return
	}

	if rest == "инфо" {
		sort.Slice(movies, func(i, j int) bool {
			return movies[i].Date.After(movies[j].Date)
		})
		s.chatService.SendWhisperToUser("🎬🍿 Всего фильмов - "+strconv.Itoa(len(movies))+
			". Последний добавленный фильм: "+movies[0].TitleRu+"&ensp;📅"+movies[0].Date.Format("02.01.2006 15:04"), msg.GetChannelId(s.userRepo.GetUserIdByWs), msg.GetUser())
		return
	}

	lang := utility.LangDetect(rest)
	sMov := s.searchMovies(movies, rest, lang)

	if len(sMov) == 0 {
		s.chatService.SendWhisperToUser("🎬🍿 Ничего не нашлось", msg.GetChannelId(s.userRepo.GetUserIdByWs), msg.GetUser())
		return
	}

	s.chatService.SendWhisperToUser(formatMsg(sMov, lang), msg.GetChannelId(s.userRepo.GetUserIdByWs), msg.GetUser())
}

func (s *LasqaService) CheckDonationAlertsStatus() {
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

	if changed := s.SetStatus(isOnline); changed {
		s.chatService.SendMessageToChannel("👀 Стример зашел в DonationAlerts", "8845069", nil)
	}
}
