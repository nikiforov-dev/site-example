package handler

import (
	"backend-example/pkg/model"
	"backend-example/pkg/websocket"
	"github.com/gin-gonic/gin"
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
	var entry model.LogEntry
	if err := c.ShouldBindJSON(&entry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	entry.StartTime = time.Now()
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
	var entry model.LogEntry
	if err := h.db.First(&entry, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Запись не найдена"})
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
