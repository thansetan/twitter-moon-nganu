package main

import (
	"embed"
	"fmt"
	"net/http"

	"github.com/thansetan/twitter-moon-nganu/cronjob"
	"github.com/thansetan/twitter-moon-nganu/handler"
	"github.com/thansetan/twitter-moon-nganu/helper"
	"github.com/thansetan/twitter-moon-nganu/util"

	"github.com/gorilla/sessions"
)

//go:embed templates/*
var templates embed.FS

//go:embed assets/*
var assets embed.FS

func main() {
	conf, err := util.LoadConfig("example.env")
	if err != nil {
		panic(err)
	}

	logger := util.NewLogger()
	httpClient := new(http.Client)
	redis, err := util.NewRedisClient(conf.RedisURL)
	if err != nil {
		panic(err)
	}

	cronjobService := cronjob.NewCronJobService(conf.CronJobAPIKey, conf.MoonEndpoint, httpClient, logger)
	sessionStore := sessions.NewCookieStore([]byte(conf.SessionKey))
	sessionStore.MaxAge(3 * 24 * 3600)
	handler := handler.New(
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

	port := helper.EnvOrDefault("PORT", "8080")
	fmt.Println("listening at port:", port)
	_ = http.ListenAndServe(fmt.Sprintf(":%s", port), mux)
}
