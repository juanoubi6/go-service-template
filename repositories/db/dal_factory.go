package db

import (
	"context"
	"database/sql"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"go-service-template/config"
	"go-service-template/log"
	"go-service-template/repositories"
	sqlTrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/database/sql"
	"regexp"
	"time"
)

// DAL = Data Access Layer
type DALFactory struct {
	locationsDBConnection *sql.DB
}

func NewDALFactory(dbConfig config.DBConfig) *DALFactory {
	conn, err := connectDB(dbConfig.LocationsDatabaseConnection, dbConfig)
	if err != nil {
		panic(err)
	}

	return &DALFactory{
		locationsDBConnection: conn,
	}
}

func (df *DALFactory) GetLocationsDB() (repositories.LocationsDB, error) {
	if df.locationsDBConnection == nil {
		return nil, errors.New("could not create LocationsDBDal because the DB connection does not exist")
	}

	return &LocationsDAL{
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
	driver := &pq.Driver{}

	// Use DataDog tracer to register all SQL queries
	sqlTrace.Register(driverName, driver, sqlTrace.WithServiceName(config.ServiceConf.AppConfig.Name), sqlTrace.WithAnalytics(true))
	dbPtr, err := sqlTrace.Open(driverName, uri)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s database: %s", driverName, err)
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
	pingDBLogger := log.GetStdLogger("pingDB")

	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("caught panic in pingDB goroutine: %v", r)
			pingDBLogger.Error("pingDB", "", err)
		}
	}()

	pingFn := func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
		defer cancel()

		err := dbPtr.PingContext(ctx)
		if err != nil {
			pingDBLogger.Error("pingDB", "", err)
		}

		time.Sleep(time.Minute * 1)
	}

	for {
		pingFn()
	}
}
