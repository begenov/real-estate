package v1

import (
	"github.com/begenov/real-estate/internal/logger"
	"github.com/begenov/real-estate/internal/model"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (h *Handler) initCollectionRoutes(api *gin.RouterGroup) {
	collection := api.Group("/collection")
	{
		public := collection.Group("/public")
		public.Use(h.cachePublicGet())
		{
			public.GET("/:key", h.getPublicCollection)
			public.GET("", h.getCollections)
		}

		authenticated := collection.Group("/private")
		authenticated.Use(h.userIdentity)
		{
			authenticated.GET("", h.getCollections)
			authenticated.GET("/:id", h.getCollectionByID)
			authenticated.POST("", h.createCollection)
			authenticated.PUT("", h.updateCollection)
			authenticated.DELETE("/:id", h.deleteCollection)
			authenticated.POST("/:id/status", h.updateCollectionStatus)

			generate := authenticated.Group("/generate")
			{
				generate.GET("/uuid/:id", h.generateUUID)
			}
		}
	}
}

func (h *Handler) getPublicCollection(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid key"})
		return
	}

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page value"})
		return
	}

	rows, err := strconv.Atoi(c.DefaultQuery("rows", "10"))
	if err != nil || rows <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid rows value"})
		return
	}

	var userId *int64

	if id, err := strconv.ParseInt(key, 10, 64); err == nil {
		collection, total, err := h.collectionService.GetCollection(c.Request.Context(), id, userId, page, rows)
		if err != nil {
			_ = c.Error(err)
			return
		}

		c.JSON(http.StatusOK, model.CollectionResponse{Collection: collection, Total: total})
		return
	}

	collection, total, err := h.collectionService.GetCollectionUUID(c, key, page, rows)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, model.CollectionResponse{Collection: collection, Total: total})
}

func (h *Handler) getCollections(c *gin.Context) {

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page value"})
		return
	}

	rows, err := strconv.Atoi(c.DefaultQuery("rows", "10"))
	if err != nil || rows <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid rows value"})
		return
	}

	search := c.Query("search")
	var searchPtr *string
	if search != "" {
		searchPtr = &search
	}

	isTemporaryStr := c.Query("is_temporary")
	var isTemporary bool
	if isTemporaryStr != "" {
		val, err := strconv.ParseBool(isTemporaryStr)
		if err == nil {
			isTemporary = val
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid is_temporary value"})
			return
		}
	}

	var userId *int64
	currentUserId, err := getUserId(c)
	if err != nil {
		logger.Warn("Failed to get user ID:", err)
	} else if currentUserId > 0 {
		userId = &currentUserId
	}

	var filter = model.CollectionFilter{
		Page:        page,
		Rows:        rows,
		Search:      searchPtr,
		IsTemporary: &isTemporary,
	}

	collections, total, err := h.collectionService.GetCollections(c.Request.Context(), &filter, userId)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, model.CollectionsResponse{Collections: collections, Total: total})
}

func (h *Handler) getCollectionByID(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid collection ID"})
		return
	}

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page value"})
		return
	}

	rows, err := strconv.Atoi(c.DefaultQuery("rows", "10"))
	if err != nil || rows <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid rows value"})
		return
	}

	currentUserId, err := getUserId(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	collection, total, err := h.collectionService.GetCollection(c.Request.Context(), id, &currentUserId, page, rows)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, model.CollectionResponse{
		Collection: collection,
		Total:      total,
	})
}

func (h *Handler) createCollection(c *gin.Context) {
	var inp model.CollectionInput
	if err := c.ShouldBind(&inp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	currentUserId, err := getUserId(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	inp.OwnerId = currentUserId

	err = h.collectionService.CreateCollection(c.Request.Context(), &inp)
	if err != nil {
		_ = c.Error(err)
		return
	}

	collection, _, err := h.collectionService.GetCollection(c.Request.Context(), inp.ID, &inp.OwnerId, 1, 1)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, collection)

	c.Status(http.StatusCreated)
}

func (h *Handler) updateCollection(c *gin.Context) {

	var inp model.CollectionInput
	if err := c.ShouldBind(&inp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	currentUserId, err := getUserId(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	inp.OwnerId = currentUserId

	err = h.collectionService.UpdateCollection(c.Request.Context(), &inp)
	if err != nil {
		_ = c.Error(err)
		return
	}

	collection, _, err := h.collectionService.GetCollection(c.Request.Context(), inp.ID, &inp.OwnerId, 1, 1)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, collection)
}

func (h *Handler) deleteCollection(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid collection ID"})
		return
	}

	currentUserId, err := getUserId(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	err = h.collectionService.DeleteCollection(c.Request.Context(), id, currentUserId)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.Status(http.StatusOK)
}

func (h *Handler) generateUUID(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid collection ID"})
		return
	}

	currentUserId, err := getUserId(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	uuid, err := h.collectionService.GenerateUUID(c.Request.Context(), id, currentUserId)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, model.CollectionUUIDResponse{UUID: uuid})
}

func (h *Handler) updateCollectionStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid collection id"})
		return
	}

	var req model.UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	currentUserId, err := getUserId(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	err = h.collectionService.UpdateStatus(c.Request.Context(), id, currentUserId, req.StatusID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.Status(http.StatusOK)
}
