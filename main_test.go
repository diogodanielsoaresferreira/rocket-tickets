package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"rocket-tickets/handler"
	"rocket-tickets/model"
	"rocket-tickets/repository"
)

func setupTestRepository(t *testing.T) repository.EventStore {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	if err := db.AutoMigrate(&model.Event{}, &model.TicketsCategory{}); err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}

	return repository.NewGormEventRepository(db)
}

func performRequest(r http.Handler, method, path string, body []byte) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)
	return resp
}

func TestPostGetAndDeleteEventFlow(t *testing.T) {
	eventRepository := setupTestRepository(t)
	gin.SetMode(gin.TestMode)
	eventHandler := handler.NewEventHandler(eventRepository)
	r := setupRouter(eventHandler)

	payload := map[string]any{
		"title":  "Summer Fest",
		"date":   "2025-07-10T19:30:00Z",
		"venue":  "City Arena",
		"artist": "The Rockets",
		"tickets": []map[string]any{
			{
				"category":  "General",
				"price":     99,
				"quantity":  100,
				"available": 100,
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	createResp := performRequest(r, http.MethodPost, "/event", body)
	if createResp.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d (body: %s)", http.StatusCreated, createResp.Code, createResp.Body.String())
	}

	var created model.Event
	if err := json.Unmarshal(createResp.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to decode create response: %v", err)
	}

	if created.ID == 0 {
		t.Fatalf("expected created event ID to be populated")
	}
	if created.Title != "Summer Fest" {
		t.Fatalf("expected title Summer Fest, got %q", created.Title)
	}
	if created.Tickets == nil || len(*created.Tickets) != 1 {
		t.Fatalf("expected one ticket category, got %+v", created.Tickets)
	}

	listResp := performRequest(r, http.MethodGet, "/event", nil)
	if listResp.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, listResp.Code)
	}

	var events []model.Event
	if err := json.Unmarshal(listResp.Body.Bytes(), &events); err != nil {
		t.Fatalf("failed to decode list response: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].ID != created.ID {
		t.Fatalf("expected event ID %d, got %d", created.ID, events[0].ID)
	}
	if events[0].Tickets == nil || len(*events[0].Tickets) != 1 {
		t.Fatalf("expected one ticket category in list response, got %+v", events[0].Tickets)
	}

	deletePath := "/event/" + strconv.FormatUint(uint64(created.ID), 10)
	deleteResp := performRequest(r, http.MethodDelete, deletePath, nil)
	if deleteResp.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, deleteResp.Code)
	}

	listAfterDelete := performRequest(r, http.MethodGet, "/event", nil)
	if listAfterDelete.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, listAfterDelete.Code)
	}

	var eventsAfterDelete []model.Event
	if err := json.Unmarshal(listAfterDelete.Body.Bytes(), &eventsAfterDelete); err != nil {
		t.Fatalf("failed to decode list-after-delete response: %v", err)
	}
	if len(eventsAfterDelete) != 0 {
		t.Fatalf("expected 0 events after delete, got %d", len(eventsAfterDelete))
	}
}

func TestPostEventWithoutTicketsDefaultsToEmptySlice(t *testing.T) {
	eventRepository := setupTestRepository(t)
	gin.SetMode(gin.TestMode)
	eventHandler := handler.NewEventHandler(eventRepository)
	r := setupRouter(eventHandler)

	payload := map[string]any{
		"title":  "Acoustic Night",
		"date":   time.Date(2025, time.August, 20, 21, 0, 0, 0, time.UTC).Format(time.RFC3339),
		"venue":  "Riverside",
		"artist": "Nora",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	resp := performRequest(r, http.MethodPost, "/event", body)
	if resp.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d (body: %s)", http.StatusCreated, resp.Code, resp.Body.String())
	}

	var created model.Event
	if err := json.Unmarshal(resp.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if created.Tickets == nil {
		t.Fatalf("expected tickets to be non-nil")
	}
	if len(*created.Tickets) != 0 {
		t.Fatalf("expected tickets to be empty, got %d entries", len(*created.Tickets))
	}
}

func TestPostEventInvalidPayload(t *testing.T) {
	eventRepository := setupTestRepository(t)
	gin.SetMode(gin.TestMode)
	eventHandler := handler.NewEventHandler(eventRepository)
	r := setupRouter(eventHandler)

	payload := []byte(`{"title":"Missing required fields"}`)
	resp := performRequest(r, http.MethodPost, "/event", payload)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, resp.Code)
	}
}

func TestDeleteEventInvalidID(t *testing.T) {
	eventRepository := setupTestRepository(t)
	gin.SetMode(gin.TestMode)
	eventHandler := handler.NewEventHandler(eventRepository)
	r := setupRouter(eventHandler)

	resp := performRequest(r, http.MethodDelete, "/event/not-a-number", nil)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, resp.Code)
	}
}

func TestDeleteEventNotFound(t *testing.T) {
	eventRepository := setupTestRepository(t)
	gin.SetMode(gin.TestMode)
	eventHandler := handler.NewEventHandler(eventRepository)
	r := setupRouter(eventHandler)

	resp := performRequest(r, http.MethodDelete, "/event/99999", nil)

	if resp.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d (body: %s)", http.StatusNotFound, resp.Code, resp.Body.String())
	}
}

func TestPostEventRejectsTicketQuantityNotGreaterThanZero(t *testing.T) {
	eventRepository := setupTestRepository(t)
	gin.SetMode(gin.TestMode)
	eventHandler := handler.NewEventHandler(eventRepository)
	r := setupRouter(eventHandler)

	payload := map[string]any{
		"title":  "Invalid Pricing Event",
		"date":   "2025-07-10T19:30:00Z",
		"venue":  "City Arena",
		"artist": "The Rockets",
		"tickets": []map[string]any{
			{
				"category":  "General",
					"price":     1,
					"quantity":  0,
					"available": 0,
				},
			},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	resp := performRequest(r, http.MethodPost, "/event", body)
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d (body: %s)", http.StatusBadRequest, resp.Code, resp.Body.String())
	}
}

func TestPostEventAllowsZeroTicketPrice(t *testing.T) {
	eventRepository := setupTestRepository(t)
	gin.SetMode(gin.TestMode)
	eventHandler := handler.NewEventHandler(eventRepository)
	r := setupRouter(eventHandler)

	payload := map[string]any{
		"title":  "Free Entry Event",
		"date":   "2025-07-10T19:30:00Z",
		"venue":  "City Arena",
		"artist": "The Rockets",
		"tickets": []map[string]any{
			{
				"category":  "General",
				"price":     0,
				"quantity":  100,
				"available": 100,
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	resp := performRequest(r, http.MethodPost, "/event", body)
	if resp.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d (body: %s)", http.StatusCreated, resp.Code, resp.Body.String())
	}
}

func TestPostEventRejectsAvailableGreaterThanQuantity(t *testing.T) {
	eventRepository := setupTestRepository(t)
	gin.SetMode(gin.TestMode)
	eventHandler := handler.NewEventHandler(eventRepository)
	r := setupRouter(eventHandler)

	payload := map[string]any{
		"title":  "Invalid Availability Event",
		"date":   "2025-07-10T19:30:00Z",
		"venue":  "City Arena",
		"artist": "The Rockets",
		"tickets": []map[string]any{
			{
				"category":  "General",
				"price":     50.5,
				"quantity":  10,
				"available": 11,
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	resp := performRequest(r, http.MethodPost, "/event", body)
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d (body: %s)", http.StatusBadRequest, resp.Code, resp.Body.String())
	}
}

func TestGetEventByIDExisting(t *testing.T) {
	eventRepository := setupTestRepository(t)
	gin.SetMode(gin.TestMode)
	eventHandler := handler.NewEventHandler(eventRepository)
	r := setupRouter(eventHandler)

	payload := map[string]any{
		"title":  "Solo Show",
		"date":   "2025-09-15T21:00:00Z",
		"venue":  "Main Hall",
		"artist": "Echo",
		"tickets": []map[string]any{
			{
				"category":  "VIP",
				"price":     120.0,
				"quantity":  20,
				"available": 20,
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}

	createResp := performRequest(r, http.MethodPost, "/event", body)
	if createResp.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d (body: %s)", http.StatusCreated, createResp.Code, createResp.Body.String())
	}

	var created model.Event
	if err := json.Unmarshal(createResp.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to decode create response: %v", err)
	}

	getPath := "/event/" + strconv.FormatUint(uint64(created.ID), 10)
	getResp := performRequest(r, http.MethodGet, getPath, nil)
	if getResp.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d (body: %s)", http.StatusOK, getResp.Code, getResp.Body.String())
	}

	var fetched model.Event
	if err := json.Unmarshal(getResp.Body.Bytes(), &fetched); err != nil {
		t.Fatalf("failed to decode get-by-id response: %v", err)
	}

	if fetched.ID != created.ID {
		t.Fatalf("expected ID %d, got %d", created.ID, fetched.ID)
	}
	if fetched.Title != "Solo Show" {
		t.Fatalf("expected title Solo Show, got %q", fetched.Title)
	}
}

func TestGetEventByIDNotFound(t *testing.T) {
	eventRepository := setupTestRepository(t)
	gin.SetMode(gin.TestMode)
	eventHandler := handler.NewEventHandler(eventRepository)
	r := setupRouter(eventHandler)

	resp := performRequest(r, http.MethodGet, "/event/99999", nil)
	if resp.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d (body: %s)", http.StatusNotFound, resp.Code, resp.Body.String())
	}
}

func TestUpdateEventReplacesTicketsAndAllowsZeroAvailable(t *testing.T) {
	eventRepository := setupTestRepository(t)
	gin.SetMode(gin.TestMode)
	eventHandler := handler.NewEventHandler(eventRepository)
	r := setupRouter(eventHandler)

	createPayload := map[string]any{
		"title":  "Launch Night",
		"date":   "2025-10-10T20:00:00Z",
		"venue":  "Sky Dome",
		"artist": "Comets",
		"tickets": []map[string]any{
			{
				"category":  "General",
				"price":     80.0,
				"quantity":  100,
				"available": 100,
			},
			{
				"category":  "VIP",
				"price":     200.0,
				"quantity":  20,
				"available": 20,
			},
		},
	}

	createBody, err := json.Marshal(createPayload)
	if err != nil {
		t.Fatalf("failed to marshal create payload: %v", err)
	}

	createResp := performRequest(r, http.MethodPost, "/event", createBody)
	if createResp.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d (body: %s)", http.StatusCreated, createResp.Code, createResp.Body.String())
	}

	var created model.Event
	if err := json.Unmarshal(createResp.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to decode create response: %v", err)
	}

	updatePayload := map[string]any{
		"title":  "Launch Night - Updated",
		"date":   "2025-10-11T20:00:00Z",
		"venue":  "Sky Dome",
		"artist": "Comets",
		"tickets": []map[string]any{
			{
				"category":  "General",
				"price":     90.0,
				"quantity":  100,
				"available": 0,
			},
		},
	}

	updateBody, err := json.Marshal(updatePayload)
	if err != nil {
		t.Fatalf("failed to marshal update payload: %v", err)
	}

	updatePath := "/event/" + strconv.FormatUint(uint64(created.ID), 10)
	updateResp := performRequest(r, http.MethodPut, updatePath, updateBody)
	if updateResp.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d (body: %s)", http.StatusOK, updateResp.Code, updateResp.Body.String())
	}

	getResp := performRequest(r, http.MethodGet, updatePath, nil)
	if getResp.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d (body: %s)", http.StatusOK, getResp.Code, getResp.Body.String())
	}

	var fetched model.Event
	if err := json.Unmarshal(getResp.Body.Bytes(), &fetched); err != nil {
		t.Fatalf("failed to decode get response: %v", err)
	}

	if fetched.Title != "Launch Night - Updated" {
		t.Fatalf("expected updated title, got %q", fetched.Title)
	}
	if fetched.Tickets == nil {
		t.Fatalf("expected tickets to be non-nil")
	}
	if len(*fetched.Tickets) != 1 {
		t.Fatalf("expected exactly 1 ticket after replacement, got %d", len(*fetched.Tickets))
	}
	if (*fetched.Tickets)[0].Available != 0 {
		t.Fatalf("expected available to be 0, got %d", (*fetched.Tickets)[0].Available)
	}
	if (*fetched.Tickets)[0].Category != "General" {
		t.Fatalf("expected remaining ticket category to be General, got %q", (*fetched.Tickets)[0].Category)
	}
}

func TestUpdateEventWithNoTicketsClearsTickets(t *testing.T) {
	eventRepository := setupTestRepository(t)
	gin.SetMode(gin.TestMode)
	eventHandler := handler.NewEventHandler(eventRepository)
	r := setupRouter(eventHandler)

	createPayload := map[string]any{
		"title":  "Clear Tickets Event",
		"date":   "2025-10-10T20:00:00Z",
		"venue":  "Sky Dome",
		"artist": "Comets",
		"tickets": []map[string]any{
			{
				"category":  "General",
				"price":     80.0,
				"quantity":  100,
				"available": 100,
			},
		},
	}

	createBody, err := json.Marshal(createPayload)
	if err != nil {
		t.Fatalf("failed to marshal create payload: %v", err)
	}

	createResp := performRequest(r, http.MethodPost, "/event", createBody)
	if createResp.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d (body: %s)", http.StatusCreated, createResp.Code, createResp.Body.String())
	}

	var created model.Event
	if err := json.Unmarshal(createResp.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to decode create response: %v", err)
	}

	updatePayload := map[string]any{
		"title":  "Clear Tickets Event",
		"date":   "2025-10-12T20:00:00Z",
		"venue":  "Sky Dome",
		"artist": "Comets",
	}

	updateBody, err := json.Marshal(updatePayload)
	if err != nil {
		t.Fatalf("failed to marshal update payload: %v", err)
	}

	updatePath := "/event/" + strconv.FormatUint(uint64(created.ID), 10)
	updateResp := performRequest(r, http.MethodPut, updatePath, updateBody)
	if updateResp.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d (body: %s)", http.StatusOK, updateResp.Code, updateResp.Body.String())
	}

	getResp := performRequest(r, http.MethodGet, updatePath, nil)
	if getResp.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d (body: %s)", http.StatusOK, getResp.Code, getResp.Body.String())
	}

	var fetched model.Event
	if err := json.Unmarshal(getResp.Body.Bytes(), &fetched); err != nil {
		t.Fatalf("failed to decode get response: %v", err)
	}

	if fetched.Tickets == nil {
		t.Fatalf("expected tickets to be non-nil")
	}
	if len(*fetched.Tickets) != 0 {
		t.Fatalf("expected tickets to be cleared, got %d", len(*fetched.Tickets))
	}
}
