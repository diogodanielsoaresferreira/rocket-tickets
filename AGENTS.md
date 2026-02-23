# Rocket Tickets Agent Guide

## Project Purpose
- Simple ticketing API in Go using Gin + GORM + SQLite.
- Core domains: `Event`, `TicketsCategory`, `Ticket`.

## Tech Stack
- Language: Go
- HTTP: Gin (`handler/event_handler.go`)
- Persistence: GORM + SQLite (`repository/repository.go`)
- Tests: Go `testing` package (`main_test.go`)

## Local Run
1. `go run main.go`
2. Server starts on `:8080`

## Test Commands
1. Run all tests: `go test ./...`
2. Run one test: `go test ./... -run TestCreateTicket -v`

## Current Routes
- `POST /event`
- `GET /event`
- `GET /event/:id`
- `PUT /event/:id`
- `DELETE /event/:id`
- `POST /event/:eventId/category/:categoryId` (sell ticket)
- `DELETE /ticket/:ticketId` (cancel ticket)
- `GET /ticket/:ticketId`

## Important Behavior Contracts
- Selling a ticket:
  - Uses guarded write (`available > 0`) inside a transaction.
  - Returns `404` when event/category pair does not exist.
  - Returns `409` when category is sold out.
- Cancelling a ticket:
  - Transactional.
  - Returns `404` when ticket does not exist.
  - Returns `409` when ticket is already cancelled.
  - Increments `available` exactly once per sold ticket.
- Event updates:
  - `PUT /event/:id` uses replacement semantics for categories.
  - If `tickets` is omitted, it is treated as empty list.

## File Map
- Entry point and routes: `main.go`
- Handlers: `handler/event_handler.go`
- Repository + DB logic: `repository/repository.go`
- Models: `model/event.go`
- Integration-style tests: `main_test.go`

## Coding Rules For Agents
- Keep sell/cancel inventory updates transactional.
- Do not reintroduce read-then-write stock decrement without guarded update.
- Preserve HTTP status code mappings already covered by tests.
- Add/adjust tests for any behavior change before considering work complete.
- Keep migrations in sync: if adding models, include them in `AutoMigrate`.
- Always run tests after code changes (`go test ./...` at minimum, or a clearly relevant subset first), then report the result.
- If tests cannot be executed, explicitly state why and what was not verified.
- After code changes, update `AGENTS.md` when routes, behaviors, architecture, or workflow expectations have changed.

## Notes
- Database file is `test.db` in project root for local runs.
- Current API uses singular resource names (`/event`, `/ticket`); preserve unless explicitly refactoring.
