package service

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/hako/branca"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/ory/dockertest"
	"log"
	"net/url"
	"testing"
)

var (
	db *sql.DB
)

func SetupTest() func() {
	//"postgresql://root@127.0.0.1:26257/ovto?sslmode=disable"
	database := "ovto"
	var err error
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	resource, err := pool.Run("cockroachdb/cockroach", "v19.1.2", []string{"POSTGRES_PASSWORD=secret", "POSTGRES_DB=" + database})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}
	resource.Expire(60)

	if err = pool.Retry(func() error {
		var err error
		db, err = sql.Open("pgx", fmt.Sprintf("postgresql://postgres:secret@localhost:%s/%s?sslmode=disable", resource.GetPort("5432/tcp"), database))
		if err != nil {
			return err
		}
		return db.Ping()
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	// When done, kill and remove the container
	tearDown :=  func() {
		err = pool.Purge(resource)
		log.Println("Purging...")
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Println("[DB online] ", db.Stats())

	return tearDown
	//m.Run()
}

func TestShouldReturnSomething(t *testing.T) {
	tearDown := SetupTest()
	defer tearDown()
	log.Println(db.Stats())
	ctx := context.TODO()

	codec := branca.NewBranca("supersecretkeyyoushouldnotcommit")
	codec.SetTTL(uint32(TokenLifeSpan.Seconds()))

	origin, err := url.Parse("http://localhost:4000")
	if err != nil || !origin.IsAbs() {
		log.Fatalf("invalid origin url: %v\n", err)
		return
	}

	s := New(db, codec, *origin)

	// Calls MenuByNameAndLanguage with mocked database connection in arguments list.
	err = s.CreateUser(ctx, "apple@gmail.com", "wdwdewffwef", "eiuwfbwieubfiuwbef")
	if err != nil {
		t.Errorf(err.Error())
	}
}

