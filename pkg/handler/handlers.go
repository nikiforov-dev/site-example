package handler

import (
	"backend-example/pkg/model"
	"backend-example/pkg/service/auth"
	"backend-example/pkg/websocket"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"net/http"
	"time"
)

type Handler struct {
	db  *gorm.DB
	ws  *websocket.WS
	loc *time.Location
}

func New(
	db *gorm.DB,
	ws *websocket.WS,
	loc *time.Location,
) *Handler {
	return &Handler{
		db:  db,
		ws:  ws,
		loc: loc,
	}
}

func (h *Handler) CreateLog(c *gin.Context) {
	login, ok := c.Get("login")
	if !ok {
		c.JSON(http.StatusBadRequest, "failed to get user login")
	}

	var user model.User
	h.db.First(&user).Where("login = ?", login)

	var entry model.LogEntry
	if err := c.ShouldBindJSON(&entry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	entry.StartTime = time.Now()
	entry.UserID = user.ID
	entry.User = user.Name
	h.db.Create(&entry)
	c.JSON(http.StatusCreated, entry)
}

func (h *Handler) FinishLog(c *gin.Context) {
	id := c.Param("id")
	var entry model.LogEntry
	if err := h.db.First(&entry, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Запись не найдена"})
		return
	}

	endTime := time.Now()
	entry.EndTime = &endTime
	entry.Duration = int64(endTime.Sub(entry.StartTime).Seconds())
	h.db.Save(&entry)
	c.JSON(http.StatusOK, entry)
}

func (h *Handler) GetLogs(c *gin.Context) {
	var entries []model.LogEntry
	h.db.Find(&entries)
	c.JSON(http.StatusOK, entries)
}

func (h *Handler) GetLogByID(c *gin.Context) {
	id := c.Param("id")
	var entry model.LogEntry
	if err := h.db.First(&entry, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Запись не найдена"})
		return
	}
	c.JSON(http.StatusOK, entry)
}

func (h *Handler) UpdateLog(c *gin.Context) {
	id := c.Param("id")

	login, ok := c.Get("login")
	if !ok {
		c.JSON(http.StatusBadRequest, "failed to get user login")
	}

	var user model.User
	h.db.First(&user).Where("login = ?", login)

	var entry model.LogEntry
	if err := h.db.First(&entry, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Запись не найдена"})
		return
	}

	if user.ID != entry.UserID {
		c.JSON(http.StatusUnauthorized, "wrong user")
		return
	}

	if err := c.ShouldBindJSON(&entry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if entry.EndTime != nil {
		entry.Duration = int64(entry.EndTime.Sub(entry.StartTime).Seconds())
	}

	h.db.Save(&entry)
	c.JSON(http.StatusOK, entry)
}

func (h *Handler) DeleteLog(c *gin.Context) {
	id := c.Param("id")
	if err := h.db.Delete(&model.LogEntry{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка при удалении"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Запись удалена"})
}

func (h *Handler) GetTotalDuration(c *gin.Context) {
	var totalDuration int
	h.db.Model(&model.LogEntry{}).Select("SUM(duration)").Scan(&totalDuration)
	c.JSON(http.StatusOK, gin.H{"total_duration": totalDuration})
}

func (h *Handler) WSHandler(c *gin.Context) {
	login, ok := c.Get("login")
	if !ok {
		c.JSON(http.StatusBadRequest, "failed to get user login")
		return
	}

	var user model.User
	h.db.First(&user).Where("login = ?", login)

	ws, err := h.ws.Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer ws.Close()
	h.ws.Clients[ws] = true
	for {
		var msg model.Message
		err := ws.ReadJSON(&msg)
		if err != nil {
			delete(h.ws.Clients, ws)
			break
		}

		msg.User = user.Name
		msg.UserID = user.ID
		h.ws.Broadcast <- msg
	}
}

func (h *Handler) GetChatMessages(c *gin.Context) {
	var messages []model.Message

	if err := h.db.Order("created_at asc").Find(&messages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось загрузить сообщения"})
		return
	}

	c.JSON(http.StatusOK, messages)
}

func (h *Handler) CreateUser(c *gin.Context) {
	var user model.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.db.Create(&user)
	c.JSON(http.StatusCreated, user)
}

func (h *Handler) GetUser(c *gin.Context) {
	type userResponse struct {
		ID    uint   `json:"id"`
		Name  string `json:"name"`
		Login string `json:"login"`
	}

	login, ok := c.Get("login")
	if !ok {
		c.JSON(http.StatusBadRequest, "failed")
	}
	login, ok = login.(string)
	if !ok {
		c.JSON(http.StatusBadRequest, "failed")
	}

	var user model.User
	h.db.First(&user).Where("login = ?", login)
	c.JSON(http.StatusCreated, userResponse{
		ID:    user.ID,
		Name:  user.Name,
		Login: user.Login,
	})
}

func (h *Handler) GetUserByID(c *gin.Context) {
	type userResponse struct {
		ID    uint   `json:"id"`
		Name  string `json:"name"`
		Login string `json:"login"`
	}

	id := c.Param("id")

	var user model.User
	h.db.First(&user).Where("id = ?", id)
	c.JSON(http.StatusCreated, userResponse{
		ID:    user.ID,
		Name:  user.Name,
		Login: user.Login,
	})
}

func (h *Handler) Login(c *gin.Context) {
	var req struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user model.User
	if err := h.db.Where("login = ?", req.Login).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не найден"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный пароль", "err": err.Error()})
		return
	}

	token, err := auth.GenerateJWT(user.Login)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка генерации токена"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}
