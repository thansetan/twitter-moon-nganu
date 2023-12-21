package handlers

import (
	"fmt"
	cronjob "go-twitter/cron-job"
	"go-twitter/utils"
	"html/template"
	"io/fs"
	"net/http"
	"time"

	oauth1Login "github.com/dghubble/gologin/oauth1"
	"github.com/dghubble/gologin/twitter"
	"github.com/dghubble/oauth1"
	twitterOAuth1 "github.com/dghubble/oauth1/twitter"
	"github.com/gorilla/sessions"
)

type HomeData struct {
	SignedIn bool
	Name     string
	History  []cronjob.JobHistory
	Err      string
}

var (
	sessionName   = "twitter-moon-session"
	sessionUserID = "userID"
	sessionJobID  = "jobID"
)

type Handler struct {
	tmpl           *template.Template
	cronJobService cronjob.CronJobService
	store          sessions.Store
	oauth1Config   *oauth1.Config
}

func New(templateFS fs.FS, cronJobService cronjob.CronJobService, store sessions.Store, conf utils.Config) Handler {
	tmpl := template.Must(template.New("templates").Funcs(template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
		"time": func(sec int64) string {
			t := time.Unix(sec, 0)
			return t.Format("02 Jan 2006 15:04:05 MST")
		},
	}).ParseFS(templateFS, "*/*.html", "*/*/*.html"))

	return Handler{
		cronJobService: cronJobService,
		tmpl:           tmpl,
		store:          store,
		oauth1Config: &oauth1.Config{
			ConsumerKey:    conf.ConsumerKey,
			ConsumerSecret: conf.ConsumerSecret,
			CallbackURL:    "http://localhost:8080/callback",
			Endpoint:       twitterOAuth1.AuthenticateEndpoint,
		},
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
		history, err := h.cronJobService.GetHistory(jobID)
		if err != nil {
			data.Err = err.Error()
		}
		data.SignedIn = true
		data.History = history
		data.Name = session.Values[sessionUserID].(string)
	}

	err := h.tmpl.ExecuteTemplate(w, "home", data)
	if err != nil {
		fmt.Println(err)
	}
}

func (h Handler) Login() http.Handler {
	return twitter.LoginHandler(h.oauth1Config, nil)
}

func (h Handler) Callback() http.Handler {
	return twitter.CallbackHandler(h.oauth1Config, http.HandlerFunc(h.createJob), nil)
}

func (h Handler) Logout(w http.ResponseWriter, r *http.Request) {
	session, _ := h.store.Get(r, sessionName)
	session.Options.MaxAge = -1
	err := session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
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

	jobID, err := h.cronJobService.CreateOrUpdate(user.IDStr, accessToken, accessTokenSecret)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session, _ := h.store.Get(r, sessionName)
	session.Values[sessionUserID] = user.IDStr
	session.Values[sessionJobID] = jobID
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}
