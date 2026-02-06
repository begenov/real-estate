package v1

import (
	"github.com/begenov/real-estate/internal/logger"
	"github.com/begenov/real-estate/internal/model"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func (h *Handler) initRealStateRoutes(api *gin.RouterGroup) {
	realEstate := api.Group("/real_estate")
	{
		public := realEstate.Group("/public")
		public.Use(h.cachePublicGet())
		{
			public.GET("", h.getRealEstates)
			public.GET("/:id", h.getRealEstateByID)
			public.GET("/total", h.getRealEstateTotal)
		}

		realEstateAuth := realEstate.Group("/private")
		realEstateAuth.Use(h.userIdentity)
		{
			realEstateAuth.GET("", h.getRealEstates)
			realEstateAuth.GET("/:id", h.getRealEstateByID)
			realEstateAuth.GET("/total", h.getRealEstateTotal)
			realEstateAuth.POST("", h.createRealEstate)
			realEstateAuth.PUT("", h.updateRealEstate)
			realEstateAuth.DELETE("/:id", h.deleteRealEstate)
			realEstateAuth.POST("/:id/status", h.updateRealEstateStatus)
		}
	}
}

func (h *Handler) getRealEstates(c *gin.Context) {
	filters, err := parseRealEstateFilters(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	var userId *int64
	currentUserId, err := getUserId(c)
	if err != nil {
		logger.Warn("Failed to get user ID:", err)
	} else if currentUserId > 0 {
		userId = &currentUserId
	}

	estates, total, err := h.realEstateService.GetRealEstates(c.Request.Context(), filters, userId)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, model.RealEstateResponse{RealEstate: estates, Total: total})
}

func (h *Handler) getRealEstateByID(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid real estate ID"})
		return
	}

	var userId *int64
	currentUserId, err := getUserId(c)
	if err != nil {
		logger.Warn("Failed to get user ID:", err)
	} else if currentUserId > 0 {
		userId = &currentUserId
	}

	estate, err := h.realEstateService.GetRealEstate(c.Request.Context(), id, userId)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, estate)
}

func (h *Handler) getRealEstateTotal(c *gin.Context) {
	filters, err := parseRealEstateFilters(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	var userId *int64
	currentUserId, err := getUserId(c)
	if err != nil {
		logger.Warn("Failed to get user ID:", err)
	} else if currentUserId > 0 {
		userId = &currentUserId
	}

	total, err := h.realEstateService.GetRealEstateTotal(c.Request.Context(), filters, userId)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, total)
}

func (h *Handler) createRealEstate(c *gin.Context) {
	var inp model.RealEstateInput
	if err := c.ShouldBind(&inp); err != nil {
		_ = c.Error(err)
		return
	}

	currentUserId, err := getUserId(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	inp.OwnerId = currentUserId

	err = h.realEstateService.CreateRealEstate(c.Request.Context(), &inp)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, inp)
}

func (h *Handler) updateRealEstate(c *gin.Context) {
	var inp model.RealEstateInput
	if err := c.ShouldBind(&inp); err != nil {
		_ = c.Error(err)
		return
	}

	currentUserId, err := getUserId(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	inp.OwnerId = currentUserId

	err = h.realEstateService.UpdateRealEstate(c.Request.Context(), &inp)
	if err != nil {
		_ = c.Error(err)
		return
	}

	estate, err := h.realEstateService.GetRealEstate(c.Request.Context(), inp.ID, &currentUserId)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, estate)
}

func (h *Handler) deleteRealEstate(c *gin.Context) {

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

	err = h.realEstateService.DeleteRealEstate(c.Request.Context(), id, currentUserId)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.Status(http.StatusOK)
}

func parseRealEstateFilters(c *gin.Context) (model.RealEstateFilter, error) {
	var filter model.RealEstateFilter

	page := c.DefaultQuery("page", "1")
	rows := c.DefaultQuery("rows", "10")

	filter.Page, _ = strconv.Atoi(page)
	filter.Rows, _ = strconv.Atoi(rows)

	if priceMinStr := c.Query("price_min"); priceMinStr != "" {
		priceMin, err := strconv.ParseFloat(priceMinStr, 64)
		if err != nil {
			return filter, err
		}
		filter.PriceMin = &priceMin
	}
	if priceMaxStr := c.Query("price_max"); priceMaxStr != "" {
		priceMax, err := strconv.ParseFloat(priceMaxStr, 64)
		if err != nil {
			return filter, model.ErrBadRequestQuery
		}
		filter.PriceMax = &priceMax
	}

	if roomsStr := c.QueryArray("room"); len(roomsStr) > 0 {
		filter.Rooms = roomsStr
	}

	if regionIDStr := c.Query("region_id"); regionIDStr != "" {
		regionID, err := strconv.Atoi(regionIDStr)
		if err != nil {
			return filter, model.ErrBadRequestQuery
		}
		filter.RegionID = &regionID
	}

	if districtIDStr := c.Query("district_id"); districtIDStr != "" {
		districtID, err := strconv.Atoi(districtIDStr)
		if err != nil {
			return filter, model.ErrBadRequestQuery
		}
		filter.DistrictID = &districtID
	}

	if typeStrs := c.QueryArray("type"); len(typeStrs) > 0 {
		for _, typeStr := range typeStrs {
			propertyType, err := strconv.Atoi(typeStr)
			if err != nil {
				return filter, model.ErrBadRequestQuery
			}
			filter.Type = append(filter.Type, propertyType)
		}
	}

	if purposeStrs := c.QueryArray("purpose"); len(purposeStrs) > 0 {
		for _, purposeStr := range purposeStrs {
			purpose, err := strconv.Atoi(purposeStr)
			if err != nil {
				return filter, model.ErrBadRequestQuery
			}
			filter.Purpose = append(filter.Purpose, purpose)
		}
	}

	sortBy := c.QueryArray("sort_by")

	for _, s := range sortBy {
		if s == "price" || s == "area" {
			filter.SortBy = append(filter.SortBy, s)
		}
	}

	sortOrder := c.QueryArray("order")
	for _, order := range sortOrder {
		if order == "asc" || order == "desc" {
			filter.SortOrder = append(filter.SortOrder, order)
		} else {
			filter.SortOrder = append(filter.SortOrder, "asc")
		}
	}

	if len(filter.SortBy) != len(filter.SortOrder) {
		return filter, model.ErrBadRequestQuery
	}

	if serialNumber := c.Query("serial_number"); serialNumber != "" {
		trimmed := strings.TrimLeft(serialNumber, "0")
		if trimmed == "" {
			return filter, model.ErrBadRequestQuery
		}

		parseInt, err := strconv.ParseInt(trimmed, 10, 64)
		if err != nil {
			return model.RealEstateFilter{}, err
		}

		filter.ID = &parseInt
	}

	if statusIDStr := c.Query("status_id"); statusIDStr != "" {
		trimmed := strings.TrimLeft(statusIDStr, "0")
		if trimmed == "" {
			return filter, model.ErrBadRequestQuery
		}

		parseInt, err := strconv.ParseInt(trimmed, 10, 64)
		if err != nil {
			return filter, model.ErrBadRequestQuery
		}

		filter.StatusID = &parseInt
	}

	return filter, nil
}

func (h *Handler) updateRealEstateStatus(c *gin.Context) {
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

	err = h.realEstateService.UpdateStatus(c.Request.Context(), id, currentUserId, req.StatusID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.Status(http.StatusOK)
}
