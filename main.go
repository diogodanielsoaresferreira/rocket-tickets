package main

import (
	"log"
	"rocket-tickets/handler"
	"rocket-tickets/repository"

	"github.com/gin-gonic/gin"
)

func setupRouter(eventHandler *handler.EventHandler) *gin.Engine {
	r := gin.Default()

	r.POST("/event", eventHandler.PostEvent)
	r.GET("/event", eventHandler.GetEvents)
	r.GET("/event/:id", eventHandler.GetEvent)
	r.PUT("/event/:id", eventHandler.UpdateEvent)
	r.DELETE("/event/:id", eventHandler.DeleteEvent)

	r.POST("/event/:eventId/category/:categoryId", eventHandler.CreateTicket)
	r.DELETE("/ticket/:ticketId", eventHandler.CancelTicket)
	r.GET("/ticket/:ticketId", eventHandler.GetTicket)

	return r
}

func main() {
	db, err := repository.Configure()
	if err != nil {
		log.Fatalf("failed to configure database: %v", err)
	}

	eventRepository := repository.NewGormEventRepository(db)
	eventHandler := handler.NewEventHandler(eventRepository)

	r := setupRouter(eventHandler)
	r.Run(":8080")
}
