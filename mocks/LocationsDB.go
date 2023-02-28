// Code generated by mockery v2.13.1. DO NOT EDIT.

package mocks

import (
	"fmt"
	domain "go-service-template/domain"

	mock "github.com/stretchr/testify/mock"

	sql "database/sql"
)

// LocationsDB is an autogenerated mock type for the LocationsDB type
type LocationsDB struct {
	mock.Mock
}

// CheckLocationNameExistence provides a mock function with given fields: ctx, name
func (_m *LocationsDB) CheckLocationNameExistence(ctx domain.ApplicationContext, name string) (bool, error) {
	ret := _m.Called(ctx, name)

	var r0 bool
	if rf, ok := ret.Get(0).(func(domain.ApplicationContext, string) bool); ok {
		r0 = rf(ctx, name)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(domain.ApplicationContext, string) error); ok {
		r1 = rf(ctx, name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CommitTx provides a mock function with given fields:
func (_m *LocationsDB) CommitTx() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateLocation provides a mock function with given fields: ctx, location
func (_m *LocationsDB) CreateLocation(ctx domain.ApplicationContext, location domain.Location) error {
	ret := _m.Called(ctx, location)

	var r0 error
	if rf, ok := ret.Get(0).(func(domain.ApplicationContext, domain.Location) error); ok {
		r0 = rf(ctx, location)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateSubLocation provides a mock function with given fields: ctx, subLocation
func (_m *LocationsDB) CreateSubLocation(ctx domain.ApplicationContext, subLocation domain.SubLocation) error {
	ret := _m.Called(ctx, subLocation)

	var r0 error
	if rf, ok := ret.Get(0).(func(domain.ApplicationContext, domain.SubLocation) error); ok {
		r0 = rf(ctx, subLocation)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Exec provides a mock function with given fields: ctx, stmt, fields
func (_m *LocationsDB) Exec(ctx domain.ApplicationContext, stmt string, fields ...interface{}) (sql.Result, error) {
	var _ca []interface{}
	_ca = append(_ca, ctx, stmt)
	_ca = append(_ca, fields...)
	ret := _m.Called(_ca...)

	var r0 sql.Result
	if rf, ok := ret.Get(0).(func(domain.ApplicationContext, string, ...interface{}) sql.Result); ok {
		r0 = rf(ctx, stmt, fields...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(sql.Result)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(domain.ApplicationContext, string, ...interface{}) error); ok {
		r1 = rf(ctx, stmt, fields...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetLocationByID provides a mock function with given fields: ctx, id
func (_m *LocationsDB) GetLocationByID(ctx domain.ApplicationContext, id string) (*domain.Location, error) {
	ret := _m.Called(ctx, id)

	var r0 *domain.Location
	if rf, ok := ret.Get(0).(func(domain.ApplicationContext, string) *domain.Location); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*domain.Location)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(domain.ApplicationContext, string) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetPaginatedLocations provides a mock function with given fields: ctx, filters
func (_m *LocationsDB) GetPaginatedLocations(ctx domain.ApplicationContext, filters domain.LocationsFilters) (domain.CursorPage[domain.Location], error) {
	ret := _m.Called(ctx, filters)

	var r0 domain.CursorPage[domain.Location]
	if rf, ok := ret.Get(0).(func(domain.ApplicationContext, domain.LocationsFilters) domain.CursorPage[domain.Location]); ok {
		r0 = rf(ctx, filters)
	} else {
		r0 = ret.Get(0).(domain.CursorPage[domain.Location])
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(domain.ApplicationContext, domain.LocationsFilters) error); ok {
		r1 = rf(ctx, filters)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Ping provides a mock function with given fields:
func (_m *LocationsDB) Ping() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RollbackTx provides a mock function with given fields:
func (_m *LocationsDB) RollbackTx() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// StartTx provides a mock function with given fields:
func (_m *LocationsDB) StartTx() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateLocation provides a mock function with given fields: ctx, location
func (_m *LocationsDB) UpdateLocation(ctx domain.ApplicationContext, location domain.Location) error {
	ret := _m.Called(ctx, location)

	var r0 error
	if rf, ok := ret.Get(0).(func(domain.ApplicationContext, domain.Location) error); ok {
		r0 = rf(ctx, location)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// WithTx provides a mock function with given fields: ctx, fn
func (_m *LocationsDB) WithTx(ctx domain.ApplicationContext, fn func(domain.ApplicationContext) error) error {
	err := _m.StartTx()
	if err != nil {
		return err
	}

	if err = fn(ctx); err != nil {
		if rollbackErr := _m.RollbackTx(); rollbackErr != nil {
			return fmt.Errorf("tx rollback failed: %w", rollbackErr)
		}

		return err
	}

	if err = _m.CommitTx(); err != nil {
		return fmt.Errorf("tx commit failed: %w", err)
	}

	return nil
}

type mockConstructorTestingTNewLocationsDB interface {
	mock.TestingT
	Cleanup(func())
}

// NewLocationsDB creates a new instance of LocationsDB. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewLocationsDB(t mockConstructorTestingTNewLocationsDB) *LocationsDB {
	mock := &LocationsDB{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
