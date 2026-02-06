package v1

import (
	"github.com/begenov/real-estate/internal/model"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (h *Handler) initAmenityRoutes(api *gin.RouterGroup) {
	amenity := api.Group("/amenity")
	amenity.Use(h.cachePublicGetWithTTL(publicPageCacheTTL))
	{
		amenity.GET("", h.getAllAmenities)
		amenity.GET("/:id", h.getAmenityByID)

		authenticated := amenity.Group("/private")
		authenticated.Use(h.userIdentity)
		{
			authenticated.POST("/", h.createAmenity)
			authenticated.PUT("/", h.updateAmenity)
			authenticated.DELETE("/:id", h.deleteAmenity)
		}
	}
}

func (h *Handler) getAllAmenities(c *gin.Context) {
	amenities, err := h.amenityService.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch amenities"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"amenities": amenities,
	})
}

func (h *Handler) getAmenityByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid amenity id"})
		return
	}

	amenity, err := h.amenityService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get amenity"})
		return
	}

	if amenity == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "amenity not found"})
		return
	}

	c.JSON(http.StatusOK, amenity)
}

func (h *Handler) createAmenity(c *gin.Context) {
	var input model.Amenity
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	err := h.amenityService.Create(c.Request.Context(), &input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create amenity"})
		return
	}

	c.JSON(http.StatusCreated, input)
}

func (h *Handler) updateAmenity(c *gin.Context) {
	var input model.Amenity
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	err := h.amenityService.Update(c.Request.Context(), &input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update amenity"})
		return
	}

	c.JSON(http.StatusOK, input)
}

func (h *Handler) deleteAmenity(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid amenity id"})
		return
	}

	err = h.amenityService.Delete(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete amenity"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "amenity deleted successfully"})
}
