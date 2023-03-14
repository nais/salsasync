package main

import (
	"context"
	"flag"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"salsasync/internal/console"
	"salsasync/internal/storage"

	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	"salsasync/internal/config"
)

var cfg = config.DefaultConfig()

func init() {
	flag.StringVar(&cfg.MetricsBindAddress, "metrics-bind-address", ":8080", "Bind address")
	flag.StringVar(&cfg.LogLevel, "log-level", "debug", "Which log level to output")
	flag.StringVar(&cfg.StorageApi, "storage-api", "", "Salsa Storage API endpoint")
	flag.StringVar(&cfg.StorageApiKey, "storage-api-key", "", "Salsa Storage API key")
	flag.StringVar(&cfg.ConsoleApiKey, "console-api-key", "", "Console API key")
}

func main() {
	parseFlags()
	setupLogger()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()

	s := storage.New(cfg.StorageApi, cfg.StorageApiKey)
	consoleApi := console.NewConfig("key1")
	teams, err := consoleApi.GetTeams(ctx)
	users, err := consoleApi.GetUsers(ctx)
	if err != nil {
		log.WithError(err).Fatal("failed to get users")
	}
	log.Debugf("Teams read %+v", teams)
	log.Debugf("Users read %+v", users)

	if err = s.SynchronizeTeamsAndUsers(teams); err != nil {
		log.WithError(err).Fatal("failed to synchronize teams and users")
	}
}

func setupLogger() {
	log.SetFormatter(&log.JSONFormatter{})
	l, err := log.ParseLevel(cfg.LogLevel)
	if err != nil {
		log.Fatal(err)
	}
	log.SetLevel(l)
}

func parseFlags() {
	err := godotenv.Load()
	if err != nil {
		log.Debugf("loading .env file %v", err)
	}

	flag.VisitAll(func(f *flag.Flag) {
		name := strings.ToUpper(strings.Replace(f.Name, "-", "_", -1))
		if value, ok := os.LookupEnv(name); ok {
			err = flag.Set(f.Name, value)
			if err != nil {
				log.Fatalf("failed setting flag from environment: %v", err)
				return
			}
		}
	})

	flag.Parse()
}
