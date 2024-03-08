package db

import (
	"context"
	"database/sql"
	"fmt"
	"go-service-template/config"
	"go-service-template/monitor"
	"go-service-template/repositories"
	"net/url"
	"regexp"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	_ "github.com/lib/pq" // side effect
	"github.com/pkg/errors"
	"github.com/uptrace/opentelemetry-go-extra/otelsql"
)

const (
	DefaultMaxOpenConns       = 10
	DefaultMaxIdleConns       = 10
	DefaultConnMaxLifetimeMin = 60
	DefaultPingSec            = 60
)

type Factory struct {
	locationsDBConnection *sql.DB
}

func NewFactory(dbConfig config.DBConfig) *Factory {
	conn, err := connectDB(dbConfig.LocationsDatabaseConnection, dbConfig)
	if err != nil {
		panic(err)
	}

	return &Factory{
		locationsDBConnection: conn,
	}
}

func (df *Factory) GetLocationsDB() (repositories.LocationsDB, error) {
	if df.locationsDBConnection == nil {
		return nil, errors.New("could not create LocationsDBDal because the DB connection does not exist")
	}

	return &LocationsRepository{
		TxDBContext:  CreateTxDBContext(df.locationsDBConnection),
		queryBuilder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}, nil
}

func connectDB(connString string, dbConfig config.DBConfig) (*sql.DB, error) {
	if connString == "" {
		return nil, errors.New("the connection string is empty")
	}

	// Pull First part of URI out
	regex := regexp.MustCompile(`(.*)://`)
	res := regex.FindAllStringSubmatch(connString, -1)

	var driverName string
	for i := range res {
		driverName = res[i][1]
	}

	uri := connString

	// Find DB name
	parsedURL, err := url.Parse(connString)
	if err != nil {
		return nil, fmt.Errorf("Error parsing connection string: %w", err)
	}
	dbName := strings.TrimPrefix(parsedURL.Path, "/")

	// Use otelsql to wrap the DB connection
	dbPtr, err := otelsql.Open(driverName, uri, otelsql.WithDBName(dbName))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s database: %w", driverName, err)
	}

	if err != nil {
		return nil, err
	}

	dbPtr.SetMaxOpenConns(dbConfig.ConnMaxIdleTime)
	if config.ServiceConf.DBConfig.ConnMaxIdleTime == 0 {
		dbPtr.SetMaxOpenConns(DefaultMaxOpenConns)
	}
	dbPtr.SetMaxIdleConns(dbConfig.MaxIdleConns)
	if config.ServiceConf.DBConfig.MaxIdleConns == 0 {
		dbPtr.SetMaxIdleConns(DefaultMaxIdleConns)
	}
	dbPtr.SetConnMaxLifetime(time.Minute * time.Duration(dbConfig.ConnMaxLifetime))
	if config.ServiceConf.DBConfig.ConnMaxLifetime == 0 {
		dbPtr.SetConnMaxLifetime(time.Minute * DefaultConnMaxLifetimeMin)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*DefaultPingSec)
	defer cancel()

	err = dbPtr.PingContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	go pingDB(dbPtr)

	return dbPtr, nil
}

func pingDB(dbPtr *sql.DB) {
	fnName := "pingDB"
	pingDBLogger := monitor.GetStdLogger(fnName)

	defer func() {
		if r := recover(); r != nil {
			pingDBLogger.Error("pingDB", "", "caught panic in pingDB goroutine", r.(error))
		}
	}()

	pingFn := func() {
		timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Second*DefaultPingSec)
		defer cancel()

		err := dbPtr.PingContext(timeoutCtx)
		if err != nil {
			pingDBLogger.Error("pingDB", "", "failed to ping DB", err)
		}

		time.Sleep(time.Minute * 1)
	}

	for {
		pingFn()
	}
}
