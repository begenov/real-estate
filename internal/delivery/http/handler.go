package http

import (
	"github.com/begenov/real-estate/internal/config"
	v1 "github.com/begenov/real-estate/internal/delivery/http/v1"
	"github.com/begenov/real-estate/internal/repository/redis"
	"github.com/begenov/real-estate/internal/service"
	"github.com/begenov/real-estate/pkg/auth"
	"github.com/begenov/real-estate/pkg/limiter"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	realEstateService service.IRealEstateService
	userService       service.IUserService
	collectionService service.ICollectionService
	minioService      service.IMinioService
	token             auth.TokenManager
	amenityService    service.IAmenityService
	locationService   service.ILocationService
	emailService      service.IEmailService
	pageService       service.IPageService
	blockService      service.IBlockService
	cache             redis.IRedisRepo
}

func NewHandler(realEstateService service.IRealEstateService, userService service.IUserService,
	collectionService service.ICollectionService, minioService service.IMinioService, token auth.TokenManager,
	amenityService service.IAmenityService, locationService service.ILocationService, emailService service.IEmailService,
	pageService service.IPageService, blockService service.IBlockService, cache redis.IRedisRepo) *Handler {
	return &Handler{
		realEstateService: realEstateService,
		userService:       userService,
		collectionService: collectionService,
		minioService:      minioService,
		token:             token,
		amenityService:    amenityService,
		locationService:   locationService,
		emailService:      emailService,
		pageService:       pageService,
		blockService:      blockService,
		cache:             cache,
	}
}

func (h *Handler) Init(cfg *config.Config) *gin.Engine {
	router := gin.Default()

	router.Use(
		gin.Recovery(),
		gin.Logger(),
		limiter.Limit(cfg.Limiter.RPS, cfg.Limiter.Burst, cfg.Limiter.TTL),
		corsMiddleware(cfg.HTTP.CORSAllowedOrigins),
		errorHandlerMiddleware,
	)

	// Init router
	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	h.initAPI(router)

	return router
}

func (h *Handler) initAPI(router *gin.Engine) {
	handlerV1 := v1.NewHandler(h.realEstateService, h.userService, h.collectionService,
		h.minioService, h.token, h.amenityService, h.locationService,
		h.emailService, h.pageService, h.blockService, h.cache)

	api := router.Group("/api")
	{
		handlerV1.Init(api)
	}
}
