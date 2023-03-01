package db

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go-service-template/monitor"
	"log"
	"testing"
)

var mockContext = monitor.CreateMockAppContext("")

type TxDBContextSuite struct {
	suite.Suite
	dbContext *TxDBContext
	sqlMock   sqlmock.Sqlmock
}

func (s *TxDBContextSuite) SetupTest() {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual), sqlmock.MonitorPingsOption(true))
	if err != nil {
		log.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}

	s.sqlMock = mock
	s.dbContext = CreateTxDBContext(db)
}

func TestTxDBContextSuite(t *testing.T) {
	suite.Run(t, new(TxDBContextSuite))
}

func (s *TxDBContextSuite) Test_WithTx_SuccessfullyWrapsTxExecutionOnSuccess() {
	s.sqlMock.ExpectBegin()
	s.sqlMock.ExpectPing()
	s.sqlMock.ExpectCommit()

	txErr := s.dbContext.WithTx(mockContext, func(fnCtx monitor.ApplicationContext) error {
		err := s.dbContext.Ping()
		if err != nil {
			return err
		}

		return nil
	})

	assert.Nil(s.T(), txErr)
	if err := s.sqlMock.ExpectationsWereMet(); err != nil {
		s.T().Errorf("there were unfulfilled expectations: %s", err)
	}
}

func (s *TxDBContextSuite) Test_WithTx_SuccessfullyWrapsTxExecutionOnFailure() {
	s.sqlMock.ExpectBegin()
	s.sqlMock.ExpectPing().WillReturnError(errors.New("ping error"))
	s.sqlMock.ExpectRollback()

	txErr := s.dbContext.WithTx(mockContext, func(fnCtx monitor.ApplicationContext) error {
		err := s.dbContext.Ping()
		if err != nil {
			return err
		}

		return nil
	})

	assert.NotNil(s.T(), txErr)
	if err := s.sqlMock.ExpectationsWereMet(); err != nil {
		s.T().Errorf("there were unfulfilled expectations: %s", err)
	}
}
