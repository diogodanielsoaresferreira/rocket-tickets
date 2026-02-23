package main

import (
	"log"
	"rocket-tickets/graph"
	"rocket-tickets/handler"
	"rocket-tickets/repository"

	gqlhandler "github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
	"github.com/vektah/gqlparser/v2/ast"
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

func setupGraphQLRoutes(r *gin.Engine, eventStore repository.EventStore) {
	graphServer := gqlhandler.New(graph.NewExecutableSchema(graph.Config{
		Resolvers: &graph.Resolver{EventStore: eventStore},
	}))
	graphServer.AddTransport(transport.Options{})
	graphServer.AddTransport(transport.GET{})
	graphServer.AddTransport(transport.POST{})
	graphServer.SetQueryCache(lru.New[*ast.QueryDocument](1000))
	graphServer.Use(extension.Introspection{})
	graphServer.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	r.GET("/", gin.WrapH(playground.Handler("GraphQL playground", "/query")))
	r.Any("/query", gin.WrapH(graphServer))
}

func main() {
	db, err := repository.Configure()
	if err != nil {
		log.Fatalf("failed to configure database: %v", err)
	}

	eventRepository := repository.NewGormEventRepository(db)
	eventHandler := handler.NewEventHandler(eventRepository)

	r := setupRouter(eventHandler)
	setupGraphQLRoutes(r, eventRepository)
	r.Run(":8080")
}
