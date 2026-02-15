package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"rocket-tickets/model"

	"rocket-tickets/repository"
)

func GetEvents(c *gin.Context) {
	var db, _ = repository.GetEvents()
	c.IndentedJSON(http.StatusOK, db)
}

func PostEvent(c *gin.Context) {
	var event model.Event

	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if event.Tickets == nil {
		event.Tickets = &[]model.TicketsCategory{}
	}

	repository.AddEvent(&event)
	c.IndentedJSON(http.StatusCreated, event)
}

func DeleteEvent(c *gin.Context) {
	idStr := c.Param("id")
	idUint64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
		return
	}
	id := uint(idUint64)

	if err := repository.DeleteEvent(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
