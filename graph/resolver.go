//go:generate go tool gqlgen generate
package graph

import "rocket-tickets/repository"

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

type Resolver struct {
	EventStore repository.EventStore
}
