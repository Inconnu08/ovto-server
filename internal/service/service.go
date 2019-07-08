package service

import (
	"database/sql"
	"net/url"

	"github.com/hako/branca"
)

// contains the core logic. You can use it to back a REST, GraphQL or RPC API :D
type Service struct {
	db     *sql.DB
	codec  *branca.Branca
	origin url.URL
}

func New(db *sql.DB, codec *branca.Branca, origin url.URL) *Service {
	return &Service{db, codec, origin}
}
