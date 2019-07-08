package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/hako/branca"
	_ "github.com/jackc/pgx/stdlib"

	"ovto/internal/handlers"
	"ovto/internal/service"
)

func main() {
	var (
		port      = env("PORT", "3000")
		originStr = env("ORIGIN", "http://localhost:"+port)
		dbURL     = env("DATABASE_URL", "postgresql://root@127.0.0.1:26257/authdemo?sslmode=disable")
		tokenKey  = env("TOKEN_KEY", "supersecretkeyyoushouldnotcommit")
	)

	origin, err := url.Parse(originStr)
	if err != nil || !origin.IsAbs() {
		log.Fatalf("invalid origin url: %v\n", err)
		return
	}

	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		log.Fatalf("cound not open db connection %v\n", err)
		return
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	if err = db.Ping(); err != nil {
		log.Fatalf("could not ping to db %v\n", err)
		return
	}

	codec := branca.NewBranca(tokenKey)
	codec.SetTTL(uint32(service.TokenLifeSpan.Seconds()))

	s := service.New(db, codec, *origin)
	h := handlers.New(s)

	//go func() {
	//	cmd := exec.Command("cockroach", "start", "--insecure")
	//	stdout, err := cmd.Output()
	//	if err != nil {
	//		println(err.Error())
	//		return
	//	}
	//
	//	print(string(stdout))
	//}()

	addr := fmt.Sprintf(":%s", port)
	log.Printf("accepting connections on port: %s\n", port)
	if err := http.ListenAndServe(addr, h); err != nil {
		log.Fatalf("could not start server: %v\n", err)
	}
}

func env(key, fallbackValue string) string {
	s, ok := os.LookupEnv(key)
	if !ok {
		return fallbackValue
	}
	return s
}

// start db
// cockroach start --insecure
//
// create table
// cat schema.sql | cockroach sql --insecure

// curl -i -X GET \
// "https://graph.facebook.com/v3.3/me?fields=id%2Cname%2Cemail%2Cbirthday%2Cpicture&access_token=EAAGKzLH7udYBAIKZA0dfE9eg0dbOVxRtkH7u1oZAIWfxy1t0pwrl7thGrpWFnmzb4zGBAN7kto5AHVu3VhYJWATHcse3zJ2DVRgIVW60SoEyRZCpFRz7EAAxKbDOHLosUCSh6EwUrAf23UNMQKqOINZCB3RV5elVcQxxqMoAcxPE9c8GBWJZB4rSDELQ3s0NSn4vJQcQ1MgZDZD"
