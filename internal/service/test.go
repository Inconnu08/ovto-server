package service

import (
	"database/sql"
	"net"
	"net/url"
	"runtime"
	"time"

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