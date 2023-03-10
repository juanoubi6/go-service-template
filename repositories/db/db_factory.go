package db

import (
	"context"
	"database/sql"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	_ "github.com/lib/pq" // side effect
	"github.com/pkg/errors"
	"go-service-template/config"
	"go-service-template/monitor"
	"go-service-template/repositories"
	"go.nhat.io/otelsql"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"regexp"
	"time"
)

// nolint
type DBFactory struct {
	locationsDBConnection *sql.DB
}

func NewDBFactory(dbConfig config.DBConfig) *DBFactory {
	conn, err := connectDB(dbConfig.LocationsDatabaseConnection, dbConfig)
	if err != nil {
		panic(err)
	}

	return &DBFactory{
		locationsDBConnection: conn,
	}
}

func (df *DBFactory) GetLocationsDB() (repositories.LocationsDB, error) {
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

	// Register the otelsql wrapper for the provided postgres driver.
	driverNameWrapper, err := otelsql.Register(driverName,
		otelsql.TraceQueryWithoutArgs(),
		otelsql.TraceRowsClose(),
		otelsql.TraceRowsAffected(),
		otelsql.WithSystem(semconv.DBSystemPostgreSQL),
	)
	if err != nil {
		return nil, err
	}

	// Use OTEL to register query traces
	dbPtr, err := sql.Open(driverNameWrapper, uri)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s database: %s", driverName, err)
	}

	if err != nil {
		return nil, err
	}

	dbPtr.SetMaxOpenConns(dbConfig.ConnMaxIdleTime)
	if config.ServiceConf.DBConfig.ConnMaxIdleTime == 0 {
		dbPtr.SetMaxOpenConns(10)
	}
	dbPtr.SetMaxIdleConns(dbConfig.MaxIdleConns)
	if config.ServiceConf.DBConfig.MaxIdleConns == 0 {
		dbPtr.SetMaxIdleConns(10)
	}
	dbPtr.SetConnMaxLifetime(time.Minute * time.Duration(dbConfig.ConnMaxLifetime))
	if config.ServiceConf.DBConfig.ConnMaxLifetime == 0 {
		dbPtr.SetConnMaxLifetime(time.Minute * 60)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	err = dbPtr.PingContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("database ping failed: %v", err.Error())
	}

	go pingDB(dbPtr)

	return dbPtr, nil
}

func pingDB(dbPtr *sql.DB) {
	fnName := "pingDB"
	pingDBLogger := monitor.GetStdLogger(fnName)
	ctx := monitor.CreateAppContextFromContext(context.Background(), fnName, "")

	defer func() {
		if r := recover(); r != nil {
			pingDBLogger.Error(ctx, "pingDB", "caught panic in pingDB goroutine", fmt.Errorf("%v", r))
		}
	}()

	pingFn := func() {
		timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Second*60)
		defer cancel()

		err := dbPtr.PingContext(timeoutCtx)
		if err != nil {
			pingDBLogger.Error(ctx, "pingDB", "failed to ping DB", err)
		}

		time.Sleep(time.Minute * 1)
	}

	for {
		pingFn()
	}
}
