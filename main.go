package main

import (
	"backend-example/pkg/handler"
	"backend-example/pkg/model"
	"backend-example/pkg/websocket"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"time"
)

var db *gorm.DB
var loc *time.Location

func initDB() error {
	var err error
	db, err = gorm.Open(sqlite.Open("logs.db"), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to get db connection")
	}
	if err := db.AutoMigrate(&model.LogEntry{}, &model.Message{}, &model.User{}); err != nil {
		return err
	}

	return nil
}

func init() {
	var err error
	loc, err = time.LoadLocation("Asia/Vladivostok")
	if err != nil {
		panic("Не удалось загрузить часовой пояс")
	}
}

func main() {
	if err := initDB(); err != nil {
		log.Fatalf("failed to init db: %v", err)
	}

	ws := websocket.New(db)
	go ws.HandleMessages()

	h := handler.New(db, ws, loc)

	r := gin.Default()

	r.StaticFile("/login", "login.html")
	r.StaticFile("/register", "register.html")
	r.POST("/user", h.CreateUser)
	r.POST("/login", h.Login)

	protected := r.Group("")
	protected.Use(handler.AuthMiddleware())

	protected.StaticFile("/", "frontend.html")
	protected.GET("/user", h.GetUser)
	protected.GET("/user/:id", h.GetUserByID)
	protected.POST("/logs", h.CreateLog)
	protected.PUT("/logs/:id/finish", h.FinishLog)
	protected.GET("/logs", h.GetLogs)
	protected.GET("/logs/:id", h.GetLogByID)
	protected.PUT("/logs/:id", h.UpdateLog)
	protected.DELETE("/logs/:id", h.DeleteLog)
	protected.GET("/logs/stats", h.GetTotalDuration)
	protected.GET("/chat/messages", h.GetChatMessages)

	protected.GET("/ws", h.WSHandler)

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Server is stopped: %v", err)
	}
}
