package repository

import (
	userrepository "avgys-gophermat/sqlc/user"
	"time"
)

type DBStore struct {
	repository *userrepository.Queries
}

const operationTimeout = 2 * time.Second

func NewRepository(repository *userrepository.Queries) *DBStore {
	return &DBStore{repository}
}
