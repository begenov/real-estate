package v1

import (
	"net/http"

	"github.com/begenov/real-estate/internal/async"
	"github.com/begenov/real-estate/internal/config"
	"github.com/begenov/real-estate/internal/repository/redis"
	"github.com/begenov/real-estate/internal/service"
	"github.com/begenov/real-estate/pkg/auth"
	"github.com/begenov/real-estate/pkg/limiter"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	realEstateService service.IRealEstateService
	userService       service.IUserService
	collectionService service.ICollectionService
	minioService      service.IMinioService
	tokenManager      auth.TokenManager
	amenityService    service.IAmenityService
	locationService   service.ILocationService
	emailService      service.IEmailService
	pageService       service.IPageService
	blockService      service.IBlockService
	cache             redis.IRedisRepo
	cacheSF           async.Singleflight
}

func NewHandler(realEstateService service.IRealEstateService,
	userService service.IUserService, collectionService service.ICollectionService,
	minioService service.IMinioService, tokenManager auth.TokenManager,
	amenityService service.IAmenityService, locationService service.ILocationService,
	emailService service.IEmailService, pageService service.IPageService,
	blockService service.IBlockService, cache redis.IRedisRepo) *Handler {
	return &Handler{
		realEstateService: realEstateService,
		userService:       userService,
		collectionService: collectionService,
		minioService:      minioService,
		tokenManager:      tokenManager,
		amenityService:    amenityService,
		locationService:   locationService,
		emailService:      emailService,
		pageService:       pageService,
		blockService:      blockService,
		cache:             cache,
	}
}

func (h *Handler) InitRouter(cfg *config.Config) *gin.Engine {
	router := gin.New()
	router.Use(
		gin.Recovery(),
		gin.Logger(),
		limiter.Limit(cfg.Limiter.RPS, cfg.Limiter.Burst, cfg.Limiter.TTL),
		corsMiddleware(cfg.HTTP.CORSAllowedOrigins),
		errorHandlerMiddleware,
	)

	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	api := router.Group("/api")
	h.Init(api)

	return router
}

func (h *Handler) Init(api *gin.RouterGroup) {
	v1 := api.Group("/v1")
	{
		h.initUserRoutes(v1)
		h.initRealStateRoutes(v1)
		h.initUploadRoutes(v1)
		h.initCollectionRoutes(v1)
		h.initAmenityRoutes(v1)
		h.initLocationRoutes(v1)
		h.initEmailRoutes(v1)
		h.initPageRoutes(v1)
		h.initBlockRoutes(v1)
	}
}
