package v1

import (
	"net/http"
	"strconv"

	"github.com/begenov/real-estate/internal/model"
	"github.com/gin-gonic/gin"
)

func (h *Handler) initTourRoutes(rg *gin.RouterGroup) {
	tour := rg.Group("/tour")
	{
		tour.POST("/public", h.submitTour)

		private := tour.Group("/private")
		private.Use(h.userIdentity)
		{
			private.GET("", h.getTourRequests)
			private.GET("/:id", h.getTourRequestByID)
			private.PUT("/:id/status", h.updateTourStatus)
		}
	}
}

func (h *Handler) submitTour(c *gin.Context) {
	var input model.TourRequestInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req, err := h.tourService.Submit(c.Request.Context(), &input)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, req)
}

func (h *Handler) getTourRequests(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page value"})
		return
	}

	rows, err := strconv.Atoi(c.DefaultQuery("rows", "10"))
	if err != nil || rows <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid rows value"})
		return
	}

	var filter model.TourRequestFilter
	filter.Page = page
	filter.Rows = rows

	if statusStr := c.Query("status_id"); statusStr != "" {
		statusID, err := strconv.Atoi(statusStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status_id"})
			return
		}
		filter.StatusID = &statusID
	}

	requests, total, err := h.tourService.GetRequests(c.Request.Context(), filter)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, model.TourRequestsResponse{
		Tours: requests,
		Total: total,
	})
}

func (h *Handler) getTourRequestByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	req, err := h.tourService.GetByID(c.Request.Context(), id)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, req)
}

func (h *Handler) updateTourStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var input model.UpdateTourStatusInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.tourService.UpdateStatus(c.Request.Context(), id, input.StatusID); err != nil {
		_ = c.Error(err)
		return
	}

	c.Status(http.StatusOK)
}
