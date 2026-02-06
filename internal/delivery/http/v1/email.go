package v1

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

func (h *Handler) initEmailRoutes(r *gin.RouterGroup) {
	email := r.Group("/email")
	{
		email.POST("/contact-form", h.sendContactForm)
		email.POST("/tour-form", h.sendTourForm)
	}
}

func (h *Handler) sendContactForm(c *gin.Context) {
	var input struct {
		Name    string `json:"name" binding:"required"`
		Phone   string `json:"phone" binding:"required"`
		Message string `json:"message" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.emailService.SendContactForm(input.Name, input.Phone, input.Message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось отправить письмо"})
		return
	}

	c.Status(http.StatusOK)
}

func (h *Handler) sendTourForm(c *gin.Context) {
	var input struct {
		Name        string    `json:"name" binding:"required"`
		Phone       string    `json:"phone" binding:"required"`
		DesiredDate time.Time `json:"desired_date" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.emailService.SendTourForm(input.Name, input.Phone, input.DesiredDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось отправить письмо"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Письмо успешно отправлено"})
}
