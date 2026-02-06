package v1

import (
	"bytes"
	"encoding/json"
	"github.com/begenov/real-estate/internal/logger"
	"github.com/begenov/real-estate/internal/model"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"strconv"
)

func (h *Handler) initUserRoutes(api *gin.RouterGroup) {
	user := api.Group("/user")
	{
		user.POST("/sign-in", h.userSignIn)
		user.POST("/auth/refresh", h.userRefresh)
		//user.POST("/sign-up", h.userCreate)

		authenticated := user.Group("/")
		authenticated.Use(h.userIdentity)
		{

			authenticated.GET("/sign-out", h.userSignOut)
			authenticated.GET("/info", h.getUserInfo)
			manager := authenticated.Group("/manager")
			{
				manager.POST("", h.userCreate)
				manager.PUT("", h.updateManager)

				manager.GET("", h.getManagers)
				manager.GET("/:id", h.getManager)
				manager.DELETE("/:id", h.deleteManager)
			}
		}
	}
}

func (h *Handler) userSignIn(c *gin.Context) {
	var inp model.UserSignInInput
	if err := c.BindJSON(&inp); err != nil {
		_ = c.Error(err)
		return
	}

	token, err := h.userService.SignIn(c.Request.Context(), inp.Username, inp.Password)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, token)
}

func (h *Handler) userCreate(c *gin.Context) {
	jsonData := c.PostForm("json")
	var inp model.UserCreateInput
	if err := json.Unmarshal([]byte(jsonData), &inp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}

	id, err := getUserId(c)
	if err != nil {
		_ = c.Error(err)
		return
	}
	inp.OwnerId = &id

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadSize)

	file, fileHeader, err := c.Request.FormFile("photo")
	if err == nil {
		defer func() {
			_ = file.Close()
		}()

		var buffer bytes.Buffer
		if _, err := io.Copy(&buffer, file); err != nil {
			logger.Error(err)
			_ = c.Error(err)
			return
		}

		contentType := http.DetectContentType(buffer.Bytes())
		if _, ex := imageTypes[contentType]; !ex {
			logger.Error(err)
			_ = c.Error(model.ErrFileNotSupported)
			return
		}

		var userPhotoInp = model.UploadInput{
			File:        bytes.NewReader(buffer.Bytes()),
			Name:        fileHeader.Filename,
			Size:        fileHeader.Size,
			ContentType: contentType,
			Bucket:      model.Avatars,
			FileType:    model.Image,
			UserID:      id,
		}

		photoURL, err := h.minioService.UploadUserPhoto(c.Request.Context(), &userPhotoInp)
		if err != nil {
			logger.Error(photoURL, err)
			return
		}

		inp.PhotoURL = &photoURL
	}

	err = h.userService.CreateUser(c.Request.Context(), &inp)
	if err != nil {
		_ = c.Error(err)
		return
	}

	user, err := h.userService.GetUser(c.Request.Context(), &model.UserFilter{Id: &inp.ID}, id)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *Handler) userSignOut(c *gin.Context) {
	tokens, err := h.parseAuthHeader(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	err = h.userService.Logout(c.Request.Context(), tokens)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}

func (h *Handler) userRefresh(c *gin.Context) {
	var inp model.RefreshInput
	if err := c.BindJSON(&inp); err != nil {
		_ = c.Error(err)
		return
	}

	token, err := h.userService.RefreshToken(c.Request.Context(), inp.Token)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, token)
}

func (h *Handler) getManagers(c *gin.Context) {
	currentUserId, err := getUserId(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	page := c.DefaultQuery("page", "1")
	rows := c.DefaultQuery("rows", "10")

	var filter = model.UsersFilter{}

	filter.Page, _ = strconv.Atoi(page)
	filter.Rows, _ = strconv.Atoi(rows)

	users, total, err := h.userService.GetUsers(c.Request.Context(), &filter, currentUserId)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, model.UsersResponse{Users: users, Total: total})
}

func (h *Handler) getManager(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	currentUserId, err := getUserId(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	var filter = model.UserFilter{
		Id: &id,
	}

	user, err := h.userService.GetUser(c.Request.Context(), &filter, currentUserId)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *Handler) updateManager(c *gin.Context) {
	jsonData := c.PostForm("json")
	var inp model.UserCreateInput
	if err := json.Unmarshal([]byte(jsonData), &inp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}

	id, err := getUserId(c)
	if err != nil {
		_ = c.Error(err)
		return
	}
	inp.OwnerId = &id

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadSize)

	file, fileHeader, err := c.Request.FormFile("photo")
	if err == nil {
		defer func() {
			_ = file.Close()
		}()

		var buffer bytes.Buffer
		if _, err := io.Copy(&buffer, file); err != nil {
			logger.Error(err)
			_ = c.Error(err)
			return
		}

		contentType := http.DetectContentType(buffer.Bytes())
		if _, ex := imageTypes[contentType]; !ex {
			logger.Error(err)
			_ = c.Error(model.ErrFileNotSupported)
			return
		}

		var userPhotoInp = model.UploadInput{
			File:        bytes.NewReader(buffer.Bytes()),
			Name:        fileHeader.Filename,
			Size:        fileHeader.Size,
			ContentType: contentType,
			Bucket:      model.Avatars,
			FileType:    model.Image,
			UserID:      id,
		}

		photoURL, err := h.minioService.UploadUserPhoto(c.Request.Context(), &userPhotoInp)
		if err != nil {
			logger.Error(photoURL, err)
			return
		}

		inp.PhotoURL = &photoURL
	}

	err = h.userService.UpdateUser(c.Request.Context(), &inp)
	if err != nil {
		_ = c.Error(err)
		return
	}

	user, err := h.userService.GetUser(c.Request.Context(), &model.UserFilter{Id: &inp.ID}, id)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *Handler) getUserInfo(c *gin.Context) {
	currentUserId, err := getUserId(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	if currentUserId == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	user, err := h.userService.GetUser(c.Request.Context(), &model.UserFilter{Id: &currentUserId}, currentUserId)
	if err != nil {
		logger.Error("h.userService.GetUser(): ", err)
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *Handler) deleteManager(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	currentUserId, err := getUserId(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	err = h.userService.DeleteUser(c.Request.Context(), currentUserId, id)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.Status(http.StatusOK)
}
