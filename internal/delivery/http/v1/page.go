package v1

import (
	"database/sql"
	"errors"
	"github.com/begenov/real-estate/internal/model"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func (h *Handler) initPageRoutes(route *gin.RouterGroup) {
	page := route.Group("/page")
	{
		public := page.Group("/public")
		public.Use(h.cachePublicGetWithTTL(publicPageCacheTTL))
		{
			public.GET("/:id", h.handleGetPage)
			public.GET("", h.handleGetPages)
		}

		private := page.Group("/private")
		private.Use(h.userIdentity)
		{
			private.POST("", h.handleCreatePage)
			private.PUT("/:id", h.handleUpdatePage)
			private.DELETE("/:id", h.handleDeletePage)
		}
	}
}

func (h *Handler) handleGetPages(c *gin.Context) {
	pages, err := h.pageService.GetPages(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to get pages",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, pages)
}

func (h *Handler) handleGetPage(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid page id",
		})
		return
	}

	page, err := h.pageService.GetPage(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "page not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "failed to get page",
				"details": err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, page)
}

func (h *Handler) handleCreatePage(c *gin.Context) {
	ctx := c.Request.Context()

	var req model.Page
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}

	if strings.TrimSpace(req.Slug) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "slug is required"})
		return
	}

	if len(req.Translation) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one translation required"})
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

	if err := h.pageService.Create(ctx, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to create page",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"page": req,
	})
}

func (h *Handler) handleUpdatePage(c *gin.Context) {
	idParam := c.Param("id")
	if idParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "missing page id",
		})
		return
	}

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid page id",
		})
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

	var page model.Page
	if err := c.ShouldBindJSON(&page); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid request body",
			"details": err.Error(),
		})
		return
	}

	page.ID = id

	err = h.pageService.Update(c.Request.Context(), &page)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to update page",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "page updated successfully",
		"page_id": page.ID,
	})
}

func (h *Handler) handleDeletePage(c *gin.Context) {
	idParam := c.Param("id")
	if idParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "missing page id",
		})
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

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid page id",
		})
		return
	}

	err = h.pageService.Delete(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to delete page",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "page deleted successfully",
		"page_id": id,
	})
}
