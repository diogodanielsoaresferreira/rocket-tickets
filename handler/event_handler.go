package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"rocket-tickets/model"

	"rocket-tickets/repository"
)

type EventHandler struct {
	repo repository.EventStore
}

func NewEventHandler(repo repository.EventStore) *EventHandler {
	return &EventHandler{repo: repo}
}

func (h *EventHandler) GetEvents(c *gin.Context) {
	var db, _ = h.repo.GetEvents()
	c.IndentedJSON(http.StatusOK, db)
}

func (h *EventHandler) PostEvent(c *gin.Context) {
	var event model.Event

	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if event.Tickets == nil {
		event.Tickets = &[]model.TicketsCategory{}
	}

	h.repo.AddEvent(&event)
	c.IndentedJSON(http.StatusCreated, event)
}

func (h *EventHandler) DeleteEvent(c *gin.Context) {
	idStr := c.Param("id")
	idUint64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
		return
	}
	id := uint(idUint64)

	if err := h.repo.DeleteEvent(id); err != nil {
		if errors.Is(err, repository.ErrEventNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
