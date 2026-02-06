package v1

import (
	"github.com/begenov/real-estate/internal/model"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (h *Handler) initLocationRoutes(rg *gin.RouterGroup) {
	location := rg.Group("/location")

	public := location.Group("/public")
	public.Use(h.cachePublicGetWithTTL(publicPageCacheTTL))
	{
		country := public.Group("/country")
		{
			country.GET("/", h.getCountries)
			country.GET("/:id", h.getCountryByID)
		}

		region := public.Group("/region")
		{
			region.GET("/", h.getRegions)
			region.GET("/:id", h.getRegionByID)
		}

		district := public.Group("/district")
		{
			district.GET("/", h.getDistricts)
			district.GET("/:id", h.getDistrictByID)
		}
	}

	private := location.Group("/private")
	{
		country := private.Group("/country")
		{
			country.POST("/", h.createCountry)
			country.PUT("/:id", h.updateCountry)
			country.DELETE("/:id", h.deleteCountry)
		}

		region := private.Group("/region")
		{
			region.POST("/", h.createRegion)
			region.PUT("/:id", h.updateRegion)
			region.DELETE("/:id", h.deleteRegion)
		}

		district := private.Group("/district")
		{
			district.POST("/", h.createDistrict)
			district.PUT("/:id", h.updateDistrict)
			district.DELETE("/:id", h.deleteDistrict)
		}
	}
}

func (h *Handler) getCountries(c *gin.Context) {
	countries, err := h.locationService.GetCountries(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, countries)
}

func (h *Handler) createCountry(c *gin.Context) {
	var input model.Country
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.locationService.CreateCountry(c.Request.Context(), &input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, input)
}

func (h *Handler) updateCountry(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var input model.Country
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input.ID = id
	if err := h.locationService.UpdateCountry(c.Request.Context(), &input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, input)
}

func (h *Handler) deleteCountry(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.locationService.DeleteCountry(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

func (h *Handler) getCountryByID(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	country, err := h.locationService.GetCountryByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if country == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Country not found"})
		return
	}

	c.JSON(http.StatusOK, country)
}

// ==================== REGION ====================

func (h *Handler) getRegions(c *gin.Context) {
	var filter model.RegionFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	regions, count, err := h.locationService.GetRegions(c.Request.Context(), &filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"regions": regions, "count": count})
}

func (h *Handler) createRegion(c *gin.Context) {
	var input model.Region
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.locationService.CreateRegion(c.Request.Context(), &input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, input)
}

func (h *Handler) updateRegion(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var input model.Region
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input.ID = id
	if err := h.locationService.UpdateRegion(c.Request.Context(), &input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, input)
}

func (h *Handler) deleteRegion(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.locationService.DeleteRegion(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

func (h *Handler) getRegionByID(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	region, err := h.locationService.GetRegionByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if region == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Region not found"})
		return
	}

	c.JSON(http.StatusOK, region)
}

// ==================== DISTRICT ====================

func (h *Handler) getDistricts(c *gin.Context) {
	var filter model.DistrictFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	districts, count, err := h.locationService.GetDistricts(c.Request.Context(), &filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"districts": districts, "count": count})
}

func (h *Handler) createDistrict(c *gin.Context) {
	var input model.District
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.locationService.CreateDistrict(c.Request.Context(), &input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, input)
}

func (h *Handler) updateDistrict(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var input model.District
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input.ID = id
	if err := h.locationService.UpdateDistrict(c.Request.Context(), &input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, input)
}

func (h *Handler) deleteDistrict(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.locationService.DeleteDistrict(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}

func (h *Handler) getDistrictByID(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	district, err := h.locationService.GetDistrictByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if district == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "District not found"})
		return
	}
	c.JSON(http.StatusOK, district)
}
