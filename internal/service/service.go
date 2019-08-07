package service

import (
	"database/sql"
	"net/url"

	"github.com/hako/branca"
)

// contains the core logic. You can use it to back a REST, GraphQL or RPC API :D
type Service struct {
	db      *sql.DB
	uCodec  *branca.Branca
	fpCodec *branca.Branca
	aCodec 	*branca.Branca
	origin  url.URL
}

func New(db *sql.DB, userCodec, foodProviderCodec, ambassadorCodec *branca.Branca, origin url.URL) *Service {
	return &Service{db, userCodec, foodProviderCodec, ambassadorCodec, origin}
}

func getRole(role int) string {
	switch role {
	case 5:
		return "Admin"
	case 10:
		 return "Owner"
	case 15:
		return "Manager"
	case 20:
		return "Supervisor"
	case 25:
		return "Waiter"
	}

	return "Waiter"
}