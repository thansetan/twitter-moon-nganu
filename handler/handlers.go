package handler

import (
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"math"
	"net/http"
	"time"

	"github.com/thansetan/twitter-moon-nganu/cronjob"
	"github.com/thansetan/twitter-moon-nganu/util"

	oauth1Login "github.com/dghubble/gologin/oauth1"
	"github.com/dghubble/gologin/twitter"
	"github.com/dghubble/oauth1"
	"github.com/gorilla/sessions"
	"github.com/redis/go-redis/v9"
)

type HomeData struct {
	Name     string
	Err      string
	History  cronjob.JobHistories
	SignedIn bool
}

var (
	sessionName     = "twitter-moon-session"
	sessionUserName = "userID"
	sessionJobID    = "jobID"
)

type Handler struct {
	tmpl           *template.Template
	cronJobService cronjob.CronJobService
	store          sessions.Store
	oauth1Config   *oauth1.Config
	redisClient    *redis.Client
	logger         *slog.Logger
}

func New(templateFS fs.FS, cronJobService cronjob.CronJobService, store sessions.Store, redisClient *redis.Client, conf util.Config, logger *slog.Logger) Handler {
	tmpl := template.Must(template.New("templates").Funcs(template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
		"time": func(sec int64) string {
			t := time.Unix(sec, 0)
			return t.Format("02 Jan 2006 15:04:05 MST")
		},
		"moon": func() string {
			julian := func(year, month, day int) float64 {
				if month <= 2 {
					year--
					month += 12
				}
				A := year / 100
				B := A / 4
				C := 2 - A + B
				E := int(365.25 * float64(year+4716))
				F := int(30.6001 * float64(month+1))
				return float64(C+day+E+F) - 1524.5
			}
			now := time.Now()
			j := julian(now.Year(), int(now.Month()), now.Day())
			p := math.Mod(j-julian(2000, 1, 6), 29.530588853)
			moonPhases := map[float64]string{
				0:        "🌑",
				1.84566:  "🌒",
				5.53588:  "🌓",
				9.22831:  "🌔",
				12.91963: "🌕",
				16.61069: "🌖",
				20.30228: "🌗",
				23.99361: "🌘",
				27.68493: "🌑",
			}
			var closestKey float64
			for k := range moonPhases {
				if k <= p {
					closestKey = max(closestKey, k)
				}
			}
			return moonPhases[closestKey]
		},
	}).ParseFS(templateFS, "*/*.html", "*/*/*.html"))

	return Handler{
		cronJobService: cronJobService,
		tmpl:           tmpl,
		store:          store,
		oauth1Config: &oauth1.Config{
			ConsumerKey:    conf.ConsumerKey,
			ConsumerSecret: conf.ConsumerSecret,
			CallbackURL:    conf.CallbackURL,
			Endpoint: oauth1.Endpoint{
				AuthorizeURL:    "https://api.x.com/oauth/authenticate",
				RequestTokenURL: "https://api.x.com/oauth/request_token",
				AccessTokenURL:  "https://api.x.com/oauth/access_token",
			},
		},
		redisClient: redisClient,
		logger:      logger,
	}
}

func (h Handler) Home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "page not found", http.StatusNotFound)
		return
	}

	var data HomeData

	session, _ := h.store.Get(r, sessionName)
	if !session.IsNew {
		jobID := session.Values[sessionJobID].(int)
		redisJobKey := fmt.Sprintf("job_%d", jobID)
		err := h.redisClient.Get(r.Context(), redisJobKey).Scan(&data.History)
		if err == redis.Nil {
			history, err := h.cronJobService.GetHistory(jobID)
			if err != nil {
				data.Err = err.Error()
			}
			err = h.redisClient.Set(r.Context(), redisJobKey, cronjob.JobHistories(history), 30*time.Minute).Err()
			if err != nil {
				h.logger.Error("redis set data error", "error", err.Error())
				http.Error(w, "something went wrong", http.StatusInternalServerError)
				return
			}
			data.History = history
		} else if err != nil {
			h.logger.Error("redis get data error", "error", err.Error())
			http.Error(w, "something went wrong", http.StatusInternalServerError)
			return
		}

		data.SignedIn = true
		data.Name = session.Values[sessionUserName].(string)
	}

	err := h.tmpl.ExecuteTemplate(w, "home", data)
	if err != nil {
		h.logger.Error("render template error", "error", err.Error())
	}
}

func (h Handler) Login() http.Handler {
	return twitter.LoginHandler(h.oauth1Config, nil)
}

func (h Handler) Callback() http.Handler {
	return twitter.CallbackHandler(h.oauth1Config, http.HandlerFunc(h.createJob), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/", http.StatusFound)
	}))
}

func (h Handler) Logout(w http.ResponseWriter, r *http.Request) {
	session, _ := h.store.Get(r, sessionName)
	session.Options.MaxAge = -1
	err := session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func (h Handler) createJob(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	accessToken, accessTokenSecret, err := oauth1Login.AccessTokenFromContext(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user, err := twitter.UserFromContext(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jobReqBody := cronjob.JobReqData{
		AccessToken:       accessToken,
		AccessTokenSecret: accessTokenSecret,
	}
	var storedJobReqBody cronjob.JobReqData
	redisUserKey := fmt.Sprintf("user_%s", user.IDStr)
	// try to get stored job data for current user
	err = h.redisClient.Get(r.Context(), redisUserKey).Scan(&storedJobReqBody)
	if err == redis.Nil { // if there's none, we create
		err = h.cronJobService.CreateOrUpdate(user.IDStr, &jobReqBody)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// also save it to redis
		err = h.redisClient.Set(r.Context(), redisUserKey, jobReqBody, 12*time.Hour).Err()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else if err != nil {
		h.logger.Error("redis get data error", "error", err.Error())
		return
	} else { // if they exists
		// check for the at and ats, if they have changed, update cron-job
		jobReqBody.JobID = storedJobReqBody.JobID
		if !jobReqBody.Eq(storedJobReqBody) {
			err = h.cronJobService.CreateOrUpdate(user.IDStr, &jobReqBody)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// save the new data to redis
			err = h.redisClient.Set(r.Context(), redisUserKey, jobReqBody, 24*time.Hour).Err()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}

	session, _ := h.store.Get(r, sessionName)
	session.Values[sessionUserName] = user.ScreenName
	session.Values[sessionJobID] = jobReqBody.JobID
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}
