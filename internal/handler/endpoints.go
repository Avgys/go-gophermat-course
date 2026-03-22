package handler

import (
	"avgys-gophermat/internal/repository"
)

type Handlers struct {
	Store repository.Repository
}

func NewHandlers(store repository.Repository) *Handlers {
	return &Handlers{Store: store}
}
