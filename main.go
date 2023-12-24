package main

import (
	"embed"
	"fmt"
	cronjob "go-twitter/cron-job"
	"go-twitter/handlers"
	"go-twitter/helpers"
	"go-twitter/utils"
	"net/http"

	"github.com/gorilla/sessions"
)

//go:embed templates/*
var templates embed.FS

//go:embed assets/*
var assets embed.FS

func main() {
	conf, err := utils.LoadConfig("example.env")
	if err != nil {
		panic(err)
	}

	logger := utils.NewLogger()
	httpClient := new(http.Client)
	redis, err := utils.NewRedisClient(conf.RedisURL)
	if err != nil {
		panic(err)
	}

	cronjobService := cronjob.NewCronJobService(conf.CronJobAPIKey, conf.MoonEndpoint, httpClient, logger)
	sessionStore := sessions.NewCookieStore([]byte(conf.SessionKey))
	sessionStore.MaxAge(3 * 24 * 3600)
	handler := handlers.New(
		templates,
		cronjobService,
		sessionStore,
		redis,
		conf,
		logger,
	)

	mux := http.NewServeMux()

	{
		mux.Handle("/assets/", http.FileServer(http.FS(assets)))
		mux.HandleFunc("/", handler.Home)
		mux.Handle("/login", handler.Login())
		mux.Handle("/callback", handler.Callback())
		mux.HandleFunc("/logout", handler.Logout)
	}

	http.ListenAndServe(fmt.Sprintf(":%s", helpers.EnvOrDefault("PORT", "8080")), mux)
}
