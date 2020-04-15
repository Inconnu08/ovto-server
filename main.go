package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/hako/branca"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/sirupsen/logrus"

	"ovto/internal/fixtures"
	"ovto/internal/handler"
	"ovto/internal/service"
)

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func run() error {
	var (
		port, _      = strconv.Atoi(env("PORT", "3000"))
		originStr    = env("ORIGIN", fmt.Sprintf("http://localhost:%d", port))
		dbURL        = env("DATABASE_URL", "postgresql://root@127.0.0.1:26257/ovto?sslmode=disable")
		userTokenKey = env("TOKEN_KEY", "supersecretkeyyoushouldnotcommit")
		fpTokenKey = env("TOKEN_KEY", "supersecretkeyyoushouldcommitokk")
		ambassadorTokenKey = env("TOKEN_KEY", "supersecretkeyyoushouldcommit111")
	)

	log := logrus.New()
	log.Formatter = &logrus.TextFormatter{
		TimestampFormat: time.RFC3339,
		FullTimestamp:   true,
		ForceColors:     true,
	}
	log.SetReportCaller(true)

	origin, err := url.Parse(originStr)
	if err != nil || !origin.IsAbs() {
		log.WithError(err).Fatal("invalid origin url:")
		return err
	}

	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		log.Fatalf("\nCould not open db connection: %v\n")
		return err
	}

	//if err := service.ValidateSchema(db); err != nil {
	//	log.Fatalf("\nFailed to validate schema: %v\n", err)
	//}

	defer func() {
		if err := db.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	if err = db.Ping(); err != nil {
		log.Fatalf("could not ping to db %v\n", err)
		return err
	}

	codec := branca.NewBranca(userTokenKey)
	codec.SetTTL(uint32(service.TokenLifeSpan.Seconds()))

	fpCodec := branca.NewBranca(fpTokenKey)
	fpCodec.SetTTL(uint32(service.TokenLifeSpan.Seconds()))

	aCodec := branca.NewBranca(ambassadorTokenKey)
	aCodec.SetTTL(uint32(service.TokenLifeSpan.Seconds()))

	s := service.New(db, codec, fpCodec, aCodec, *origin)

	fixtures.PopulateFoodProvider(s)

	server := http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           handler.New(s),
		ReadHeaderTimeout: time.Second * 5,
		ReadTimeout:       time.Second * 15,
	}

	errs := make(chan error, 2)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt, os.Kill)

		<-quit

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			errs <- fmt.Errorf("could not shutdown server: %v", err)
			return
		}

		errs <- ctx.Err()
	}()

	go func() {
		log.Printf("accepting connections on port %d\n", port)
		log.Printf("starting server at %s\n", origin)
		if err = server.ListenAndServe(); err != http.ErrServerClosed {
			errs <- fmt.Errorf("could not listen and serve: %v", err)
			return
		}

		errs <- nil
	}()

	return <-errs
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
// cockroach start --insecure --host localhost
//
// create table
// cat schema.sql | cockroach sql --insecure

// curl -i -X GET \
// "https://graph.facebook.com/v3.3/me?fields=id%2Cname%2Cemail%2Cbirthday%2Cpicture&access_token=EAAGKzLH7udYBAIKZA0dfE9eg0dbOVxRtkH7u1oZAIWfxy1t0pwrl7thGrpWFnmzb4zGBAN7kto5AHVu3VhYJWATHcse3zJ2DVRgIVW60SoEyRZCpFRz7EAAxKbDOHLosUCSh6EwUrAf23UNMQKqOINZCB3RV5elVcQxxqMoAcxPE9c8GBWJZB4rSDELQ3s0NSn4vJQcQ1MgZDZD"

// curl -H "Accept: text/event-stream" -H "Authorization: Bearer ATj7V7RiWgP8eX5WEd0gKdzfE6QTuPxoLUz0RbVYg55wyfv5gF708C4HXw49rTyJD1BI2e4JJ1Nx1876JM4qf" http://localhost:3000/api/restaurants/00e1630a-8c56-47f8-8dbc-56e17953941c/orders