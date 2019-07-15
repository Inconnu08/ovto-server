package service

import (
	"context"
	"database/sql"
	"github.com/google/go-cmp/cmp"
	"net"
	"net/url"
	"runtime"
	"testing"
	"time"

	"github.com/hako/branca"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/stdlib"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/ory/dockertest"
	"github.com/ory/dockertest/docker"
	"github.com/sirupsen/logrus"
)

var (
	log *logrus.Logger

	pgURL *url.URL
)

// before
func SetupTest() func() {
	//"postgresql://root@127.0.0.1:26257/db?sslmode=disable"
	log = logrus.New()
	log.Formatter = &logrus.TextFormatter{
		TimestampFormat: time.RFC3339,
		FullTimestamp:   true,
		ForceColors:     true,
	}

	pgURL = &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword("myuser", "mypass"),
		Path:   "post",
	}
	q := pgURL.Query()
	q.Add("sslmode", "disable")
	pgURL.RawQuery = q.Encode()

	pool, err := dockertest.NewPool("")
	if err != nil {
		log.WithError(err).Fatal("Could not connect to docker")
	}

	pw, _ := pgURL.User.Password()
	runOpts := dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "latest",
		Env: []string{
			"POSTGRES_USER=" + pgURL.User.Username(),
			"POSTGRES_PASSWORD=" + pw,
			"POSTGRES_DB=" + pgURL.Path,
		},
	}

	resource, err := pool.RunWithOptions(&runOpts)
	if err != nil {
		log.WithError(err).Fatal("Could not start postgres container")
	}

	pgURL.Host = resource.Container.NetworkSettings.IPAddress

	// Docker layer network is different on Mac
	if runtime.GOOS == "darwin" {
		pgURL.Host = net.JoinHostPort(resource.GetBoundIP("5432/tcp"), resource.GetPort("5432/tcp"))
	}

	logWaiter, err := pool.Client.AttachToContainerNonBlocking(docker.AttachToContainerOptions{
		Container:    resource.Container.ID,
		OutputStream: log.Writer(),
		ErrorStream:  log.Writer(),
		Stderr:       true,
		Stdout:       true,
		Stream:       true,
	})
	if err != nil {
		log.WithError(err).Fatal("Could not connect to postgres container log output")
	}
	defer func() {
		err = logWaiter.Close()
		if err != nil {
			log.WithError(err).Error("Could not close container log")
		}
		err = logWaiter.Wait()
		if err != nil {
			log.WithError(err).Error("Could not wait for container log to close")
		}
	}()

	pool.MaxWait = 10 * time.Second
	err = pool.Retry(func() error {
		db, err := sql.Open("postgres", pgURL.String())
		if err != nil {
			return err
		}
		return db.Ping()
	})
	if err != nil {
		log.WithError(err).Fatal("Could not connect to postgres server")
	}

	// When done, kill and remove the container
	tearDown := func() {
		err = pool.Purge(resource)
		if err != nil {
			log.Fatal(err)
		}
	}

	return tearDown
}

func TestService_CreateUser(t *testing.T) {
	var tt = []struct {
		Label     string
		Condition string
		Want      error
		Email     string
		Fullname  string
		Password  string
	}{
		// success Condition
		{Label: "Test should create user successfully", Condition: "success", Want: nil, Email: "johndoe@gmail.com", Fullname: "John Snow", Password: "ShesMyQueen"},
		// error Condition
		{Label: "Test should if Email taken", Condition: "fail", Want: ErrEmailTaken, Email: "johndoe@gmail.com", Fullname: "John Snow", Password: "ShesMyQueen"},
	}

	tearDown := SetupTest()
	defer tearDown()

	ctx := context.TODO()

	codec := branca.NewBranca("supersecretkeyyoushouldnotcommit")
	codec.SetTTL(uint32(TokenLifeSpan.Seconds()))

	c, err := pgx.ParseURI(pgURL.String())
	if err != nil {
		log.Fatalf(err.Error())
	}

	db := stdlib.OpenDB(c)

	if err := ValidateSchema(db); err != nil {
		log.Fatalf("failed to validate schema: %v\n", err)
	}

	s := New(db, codec, url.URL{})

	for _, test := range tt {
		t.Run(test.Label, func(t *testing.T) {
			Got := s.CreateUser(ctx, test.Email, test.Fullname, test.Password)
			if test.Condition == "success" {
				if Got != nil {
					t.Error("Got:", Got, "| Want:", test.Want)
				}
			} else {
				if !cmp.Equal(Got.Error(), test.Want.Error()) {
					t.Error("Got:", Got, "| Want:", test.Want)
				}
			}
		})
	}

}
