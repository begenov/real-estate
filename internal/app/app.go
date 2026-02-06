package app

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/begenov/real-estate/internal/async"
	"github.com/begenov/real-estate/internal/config"
	httpV1 "github.com/begenov/real-estate/internal/delivery/http"
	"github.com/begenov/real-estate/internal/logger"
	"github.com/begenov/real-estate/internal/repository/minio"
	"github.com/begenov/real-estate/internal/repository/postgres"
	"github.com/begenov/real-estate/internal/repository/redis"
	"github.com/begenov/real-estate/internal/server"
	"github.com/begenov/real-estate/internal/service"
	"github.com/begenov/real-estate/pkg/auth"
	"github.com/begenov/real-estate/pkg/db"
	"github.com/begenov/real-estate/pkg/hash"
	"github.com/begenov/real-estate/pkg/smtp"
	"github.com/begenov/real-estate/pkg/translate"
)

const timeout = 10 * time.Second

func Run(cfg *config.Config) error {

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	redisClient, err := db.NewRedisClient(ctx, cfg.Redis)
	if err != nil {
		logger.Error("db.NewRedisClient(): ", err)
		return err
	}
	postgresDB, err := db.NewDatabase(ctx, cfg.Postgres.Driver, cfg.Postgres.DSN)
	if err != nil {
		logger.Error("db.NewDatabase(): ", err)
		return err
	}

	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.Error("postgresDB.Close(): ", err)
			return
		}
	}(postgresDB)

	clientMinio, err := db.NewMinioClient(ctx, cfg.Minio)
	if err != nil {
		return err
	}

	translateClient, err := translate.NewTranslateClient(ctx, cfg.Translate.CredentialsPath)
	if err != nil {
		return err
	}

	defer func() {
		err = translateClient.Close()
		if err != nil {
			logger.Error("translateClient.Close(): ", err)
			return
		}
	}()

	hashPassword := hash.NewHash()

	token, err := auth.NewManager(cfg.JWT.AccessSigningKey, cfg.JWT.RefreshSigningKey, cfg.JWT.AccessTokenTTL, cfg.JWT.RefreshTokenTTL)
	if err != nil {
		return err
	}

	smtpSender, err := smtp.NewSMTPSender(cfg.SMTP.From, cfg.SMTP.Password, cfg.SMTP.Host, cfg.SMTP.Port)
	if err != nil {
		logger.Error(err)
		return err
	}

	//Repo
	userRepo := postgres.NewUserRepo(postgresDB)
	realEstateRepo := postgres.NewRealEstateRepo(postgresDB, clientMinio)
	collectionRepo := postgres.NewCollectionRepo(postgresDB)
	txRepo := postgres.NewTxRepo(postgresDB)
	fileRepo := postgres.NewFileRepo(postgresDB)
	exchangeRateRepo := postgres.NewExchangeRateRepo(postgresDB)

	redisRepo := redis.NewRedisRepo(redisClient)

	minioRepo := minio.NewMinioRepo(clientMinio)
	amenityRepo := postgres.NewAmenityRepo(postgresDB)
	locationRepo := postgres.NewLocationRepo(postgresDB)
	pageRepo := postgres.NewPageRepo(postgresDB)
	blockRepo := postgres.NewBlockRepo(postgresDB)

	//Service
	// TODO: Получать путь к логотипу через config/env
	translateService := service.NewTranslateService(translateClient)
	exchangeRateService := service.NewExchangeRateService(exchangeRateRepo, cfg.Exchange.APIKey)
	imageService := service.NewImageService(cfg.Watermark.Path)
	minioService := service.NewMinioService(
		fileRepo,
		minioRepo,
		imageService,
		cfg.Minio.Endpoint,
		cfg.Environment,
		cfg.Async.MinioReprocessWorkers,
		cfg.Async.MinioReprocessQueue,
		cfg.Minio.Buckets...,
	)
	amenityService := service.NewAmenityService(amenityRepo)

	locationService := service.NewLocationService(locationRepo)
	realEstateService := service.NewRealEstateService(
		realEstateRepo,
		userRepo,
		txRepo,
		minioService,
		amenityService,
		locationService,
		exchangeRateService,
		translateService,
		cfg.Async.TranslateConcurrency,
	)
	userService := service.NewUserService(hashPassword, token, userRepo, redisRepo, txRepo)
	collectionService := service.NewCollectionService(collectionRepo, realEstateRepo, userRepo, txRepo, exchangeRateService, translateService)
	emailService := service.NewEmailService(smtpSender, cfg.EmailConfig, cfg.SMTP.From, blockRepo)
	pageService := service.NewPageService(pageRepo)
	blockService := service.NewBlockService(blockRepo, translateService)

	// handler
	handler := httpV1.NewHandler(realEstateService, userService, collectionService, minioService, token, amenityService, locationService, emailService, pageService, blockService, redisRepo)

	srv := server.NewServer(cfg, handler.Init(cfg))

	appCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	group, groupCtx := async.WithContext(appCtx)
	group.Go(async.WithTiming("exchange_rate_updater", func(ctx context.Context) error {
		exchangeRateService.StartUpdater(ctx)
		return nil
	}))
	group.Go(async.WithTiming("http_server", func(ctx context.Context) error {
		if err := srv.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		return nil
	}))

	<-groupCtx.Done()
	logger.Info("Shutting down...")

	const timeout = 5 * time.Second
	shutdownCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := srv.Stop(shutdownCtx); err != nil {
		logger.Error("failed to stop server: %v", err)
	}

	if err := group.Wait(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("background error: %v", err)
	}
	return nil
}
