package v1

import (
	"database/sql"
	"errors"
	"github.com/begenov/real-estate/internal/model"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (h *Handler) initBlockRoutes(route *gin.RouterGroup) {
	block := route.Group("/block")
	{
		public := block.Group("/public")
		public.Use(h.cachePublicGetWithTTL(publicPageCacheTTL))
		{
			public.GET("/:id", h.handleGetBlock)
			public.GET("/list/:page_id", h.handleGetBlocks)
		}

		private := block.Group("/private")
		private.Use(h.userIdentity)
		{
			private.POST("", h.handleCreateBlock)
			private.PUT("/:id", h.handleUpdateBlock)
		}
	}

}

func (h *Handler) handleGetBlocks(c *gin.Context) {
	pageIDStr := c.Param("page_id")
	if pageIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing block id"})
		return
	}

	pageID, err := strconv.ParseInt(pageIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page_id"})
		return
	}

	blocks, err := h.blockService.GetBlocks(c.Request.Context(), pageID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to get blocks",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"blocks": blocks,
	})
}

func (h *Handler) handleGetBlock(c *gin.Context) {
	idParam := c.Param("id")
	if idParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing block id"})
		return
	}

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid block id"})
		return
	}

	block, err := h.blockService.GetBlock(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "block not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get block", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"block": block})
}

func (h *Handler) handleCreateBlock(c *gin.Context) {
	var req model.Block
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}

	if req.PageID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "page_id is required"})
		return
	}

	currentUserId, err := getUserId(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	user, err := h.userService.GetUser(c.Request.Context(), &model.UserFilter{
		Id: &currentUserId,
	}, currentUserId)
	if err != nil {
		_ = c.Error(err)
		return
	}

	if !model.HasRoles(user.Roles, model.Role_Admin) {
		c.JSON(http.StatusForbidden, gin.H{})
		return
	}

	if err := h.blockService.Create(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create block", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "block created successfully",
		"block":   req,
	})
}

func (h *Handler) handleUpdateBlock(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request id", "details": err.Error()})
		return
	}

	currentUserId, err := getUserId(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	user, err := h.userService.GetUser(c.Request.Context(), &model.UserFilter{
		Id: &currentUserId,
	}, currentUserId)
	if err != nil {
		_ = c.Error(err)
		return
	}

	if !model.HasRoles(user.Roles, model.Role_Admin) {
		c.JSON(http.StatusForbidden, gin.H{})
		return
	}

	var req model.Block
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}

	req.ID = id

	if err := h.blockService.Update(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update block", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"block": req})
}
