package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/borisjacquot/juno/internal/bot"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

func main() {

	logger := setupLogging()

	logger.Info("Starting Juno bot...")
	// load env vars
	if err := godotenv.Load(); err != nil {
		logger.Info("No .env file found, using system env vars instead")
	}

	// setup env vars
	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		logger.Fatal("DISCORD_TOKEN is not set")
	}

	overfastURL := os.Getenv("OVERFAST_API_URL")
	if overfastURL == "" {
		overfastURL = "https://overfast-api.tekrop.fr"
		logger.WithField("url", overfastURL).Info("OVERFAST_API_URL is not set, using default")
	}

	logger.WithFields(log.Fields{
		"token":        token[:2] + "******",
		"overfast_url": overfastURL,
	}).Info("Environment variables loaded")

	// init bot
	b, err := bot.NewBot(token, overfastURL, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to create bot instance")
	}

	// start bot
	if err := b.Start(); err != nil {
		logger.WithError(err).Fatal("Failed to start bot")
	}
	defer b.Stop()

	logger.Info("Juno bot is started. Press CTRL+C to gracefully stop.")

	// wait for interrupt signal
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	logger.Info("Stop signal received, shutting down Juno bot...")
}

func setupLogging() *log.Logger {
	// setup logs
	logger := log.New()

	logger.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	logger.SetOutput(os.Stdout)

	// handle loglevel
	logLevel := os.Getenv("LOG_LEVEL")
	switch logLevel {
	case "debug":
		logger.SetLevel(log.DebugLevel)
	case "warn":
		logger.SetLevel(log.WarnLevel)
	case "error":
		logger.SetLevel(log.ErrorLevel)
	default:
		logger.SetLevel(log.InfoLevel)
	}

	return logger
}
