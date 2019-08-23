package service

import (
	"database/sql"
	"net/url"
	"sync"

	"github.com/hako/branca"
)

// contains the core logic. You can use it to back a REST, GraphQL or RPC API :D
type Service struct {
	db           *sql.DB
	uCodec       *branca.Branca
	fpCodec      *branca.Branca
	aCodec       *branca.Branca
	origin       url.URL
	orderClients sync.Map
}

func New(db *sql.DB, userCodec, foodProviderCodec, ambassadorCodec *branca.Branca, origin url.URL) *Service {
	return &Service{db:db, uCodec: userCodec, fpCodec:foodProviderCodec, aCodec:ambassadorCodec, origin:origin}
}
