package graph

import (
	"github.com/gin-gonic/gin/binding"
	"rocket-tickets/model"
)

func validateEventPayload(event *model.Event) error {
	return binding.Validator.ValidateStruct(event)
}

