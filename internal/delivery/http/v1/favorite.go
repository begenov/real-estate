package v1

import (
	"net/http"
	"strconv"

	"github.com/begenov/real-estate/internal/model"
	"github.com/gin-gonic/gin"
)

func (h *Handler) initFavoriteRoutes(rg *gin.RouterGroup) {
	favorites := rg.Group("/favorites")
	favorites.Use(h.userIdentity)
	{
		favorites.POST("", h.addFavorite)
		favorites.DELETE("/:id", h.removeFavorite)
		favorites.GET("", h.getFavorites)
	}
}

func (h *Handler) addFavorite(c *gin.Context) {
	var input model.FavoriteInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := getUserId(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	if err := h.favoriteService.Add(c.Request.Context(), userID, input.RealEstateID); err != nil {
		_ = c.Error(err)
		return
	}

	c.Status(http.StatusCreated)
}

func (h *Handler) removeFavorite(c *gin.Context) {
	realEstateID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || realEstateID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	userID, err := getUserId(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	if err := h.favoriteService.Remove(c.Request.Context(), userID, realEstateID); err != nil {
		_ = c.Error(err)
		return
	}

	c.Status(http.StatusOK)
}

func (h *Handler) getFavorites(c *gin.Context) {
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

	userID, err := getUserId(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	favorites, total, err := h.favoriteService.GetFavorites(c.Request.Context(), userID, page, rows)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, model.FavoritesResponse{
		Favorites: favorites,
		Total:     total,
	})
}
