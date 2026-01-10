package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	migratePostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var repoSuite Repo

func TestMain(m *testing.M) {
	ctx := context.Background()

	terminateContainerAndFatal := func(c testcontainers.Container, err error) {
		c.Terminate(ctx)
		log.Fatal(err)
	}

	dbName := "offer_eligibility"
	dbUser := "user"
	dbPass := "pass"

	env := map[string]string{
		"POSTGRES_DB":       dbName,
		"POSTGRES_USER":     dbUser,
		"POSTGRES_PASSWORD": dbPass,
	}

	c, err := testcontainers.Run(
		ctx,
		"postgres:16",
		testcontainers.WithExposedPorts("5432/tcp"),
		testcontainers.WithEnv(env),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("5432/tcp"),
		),
	)
	if err != nil {
		log.Fatal(err)
	}

	endpoint, err := c.Endpoint(ctx, "")
	if err != nil {
		terminateContainerAndFatal(c, err)
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", dbUser, dbPass, endpoint, dbName)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		terminateContainerAndFatal(c, err)
	}

	pingErr := db.Ping()
	if pingErr != nil {
		terminateContainerAndFatal(c, pingErr)
	}

	driver, err := migratePostgres.WithInstance(db, &migratePostgres.Config{})
	if err != nil {
		terminateContainerAndFatal(c, err)
	}
	mig, err := migrate.NewWithDatabaseInstance("file://migrations", "postgres", driver)
	if err != nil {
		terminateContainerAndFatal(c, err)
	}
	if err = mig.Up(); err != nil {
		terminateContainerAndFatal(c, err)
	}

	repoSuite = NewRepo(db)

	code := m.Run()

	if err = mig.Down(); err != nil {
		terminateContainerAndFatal(c, err)
	}

	c.Terminate(ctx)
	os.Exit(code)
}
