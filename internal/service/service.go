package service

import (
	"database/sql"
	"net/url"

	"github.com/hako/branca"
	"github.com/matryer/vice"
)

// contains the core logic. You can use it to back a REST, GraphQL or RPC API :D
type Service struct {
	db        *sql.DB
	transport vice.Transport
	uCodec    *branca.Branca
	fpCodec   *branca.Branca
	aCodec    *branca.Branca
	origin    url.URL
}

func New(db *sql.DB, transport vice.Transport, userCodec, foodProviderCodec, ambassadorCodec *branca.Branca, origin url.URL) *Service {
	return &Service{db, transport, userCodec, foodProviderCodec, ambassadorCodec, origin}
}
