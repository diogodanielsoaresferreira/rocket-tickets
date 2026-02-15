package main

import (
	"rocket-tickets/handler"
	"rocket-tickets/repository"

	"github.com/gin-gonic/gin"
)

func setupRouter() *gin.Engine {
	r := gin.Default()

	r.POST("/event", handler.PostEvent)
	r.GET("/event", handler.GetEvents)
	r.DELETE("/event/:id", handler.DeleteEvent)

	return r
}

func main() {
	repository.Configure()
	r := setupRouter()
	r.Run(":8080")
}
