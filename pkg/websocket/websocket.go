package websocket

import (
	"backend-example/pkg/model"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
	"net/http"
	"time"
)

type WS struct {
	db        *gorm.DB
	Upgrader  websocket.Upgrader
	Clients   map[*websocket.Conn]bool
	Broadcast chan model.Message
}

func New(db *gorm.DB) *WS {
	return &WS{
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		Clients:   make(map[*websocket.Conn]bool),
		Broadcast: make(chan model.Message),

		db: db,
	}
}

// HandleMessages Рассылка сообщений всем клиентам
func (w *WS) HandleMessages() {
	for {
		msg := <-w.Broadcast
		msg.CreatedAt = time.Now()
		w.db.Create(&msg)

		for client := range w.Clients {
			err := client.WriteJSON(msg)
			if err != nil {
				client.Close()
				delete(w.Clients, client)
			}
		}
	}
}
