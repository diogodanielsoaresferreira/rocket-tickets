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

func (h *EventHandler) GetEvent(c *gin.Context) {
	idStr := c.Param("id")
	idUint64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
		return
	}
	id := uint(idUint64)

	event, err := h.repo.GetEvent(id)
	if err != nil {
		if errors.Is(err, repository.ErrEventNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, event)
}

func (h *EventHandler) UpdateEvent(c *gin.Context) {
	idStr := c.Param("id")
	idUint64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
		return
	}
	id := uint(idUint64)

	var event model.Event
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if event.Tickets == nil {
		event.Tickets = &[]model.TicketsCategory{}
	}

	event.ID = id

	if err := h.repo.UpdateEvent(id, &event); err != nil {
		if errors.Is(err, repository.ErrEventNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, event)
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

func (h *EventHandler) CreateTicket(c *gin.Context) {
	eventIdStr := c.Param("eventId")
	eventIdUint64, err := strconv.ParseUint(eventIdStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event ID"})
		return
	}
	eventId := uint(eventIdUint64)

	categoryIdStr := c.Param("categoryId")
	categoryIdUint64, err := strconv.ParseUint(categoryIdStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category ID"})
		return
	}
	categoryId := uint(categoryIdUint64)

	ticket, err := h.repo.CreateTicket(eventId, categoryId)
	if err != nil {
		if errors.Is(err, repository.ErrEventNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, repository.ErrNoTicketsAvailable) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusCreated, ticket)
}

func (h *EventHandler) CancelTicket(c *gin.Context) {
	ticketIdStr := c.Param("ticketId")
	ticketIdUint64, err := strconv.ParseUint(ticketIdStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ticket ID"})
		return
	}
	ticketId := uint(ticketIdUint64)

	if err := h.repo.CancelTicket(ticketId); err != nil {
		if errors.Is(err, repository.ErrTicketNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, repository.ErrTicketAlreadyCancelled) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *EventHandler) GetTicket(c *gin.Context) {
	ticketIdStr := c.Param("ticketId")
	ticketIdUint64, err := strconv.ParseUint(ticketIdStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ticket ID"})
		return
	}
	ticketId := uint(ticketIdUint64)

	ticket, err := h.repo.GetTicket(ticketId)
	if err != nil {
		if errors.Is(err, repository.ErrTicketNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, ticket)
}
