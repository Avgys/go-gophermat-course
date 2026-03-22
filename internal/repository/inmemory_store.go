package repository

import (
	"sync"

	"avgys-gophermat/internal/model"

	"github.com/samber/lo"
)

type storage map[string]*model.DBURL

type InMemoryStore struct {
	shortURLToModel storage
	mux             sync.RWMutex
}

func NewInMemoryStore(initData []*model.DBURL) *InMemoryStore {

	mappedURLs := lo.Associate(initData, func(item *model.DBURL) (string, *model.DBURL) { return item.ShortURL, item })
	s := &InMemoryStore{shortURLToModel: mappedURLs}

	return s
}
