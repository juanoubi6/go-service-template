// Code generated by mockery v2.13.1. DO NOT EDIT.

package mocks

import (
	domain "go-service-template/domain"
	dto "go-service-template/domain/dto"

	mock "github.com/stretchr/testify/mock"
)

// ILocationService is an autogenerated mock type for the ILocationService type
type ILocationService struct {
	mock.Mock
}

// CreateLocation provides a mock function with given fields: ctx, newLocationData
func (_m *ILocationService) CreateLocation(ctx domain.ApplicationContext, newLocationData dto.CreateLocationRequest) (domain.Location, error) {
	ret := _m.Called(ctx, newLocationData)

	var r0 domain.Location
	if rf, ok := ret.Get(0).(func(domain.ApplicationContext, dto.CreateLocationRequest) domain.Location); ok {
		r0 = rf(ctx, newLocationData)
	} else {
		r0 = ret.Get(0).(domain.Location)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(domain.ApplicationContext, dto.CreateLocationRequest) error); ok {
		r1 = rf(ctx, newLocationData)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetLocationByID provides a mock function with given fields: ctx, id
func (_m *ILocationService) GetLocationByID(ctx domain.ApplicationContext, id string) (*domain.Location, error) {
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
func (_m *ILocationService) GetPaginatedLocations(ctx domain.ApplicationContext, filters domain.LocationsFilters) (domain.CursorPage[domain.Location], error) {
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

// UpdateLocation provides a mock function with given fields: ctx, updatedLocationData
func (_m *ILocationService) UpdateLocation(ctx domain.ApplicationContext, updatedLocationData dto.UpdateLocationRequest) (domain.Location, error) {
	ret := _m.Called(ctx, updatedLocationData)

	var r0 domain.Location
	if rf, ok := ret.Get(0).(func(domain.ApplicationContext, dto.UpdateLocationRequest) domain.Location); ok {
		r0 = rf(ctx, updatedLocationData)
	} else {
		r0 = ret.Get(0).(domain.Location)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(domain.ApplicationContext, dto.UpdateLocationRequest) error); ok {
		r1 = rf(ctx, updatedLocationData)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewILocationService interface {
	mock.TestingT
	Cleanup(func())
}

// NewILocationService creates a new instance of ILocationService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewILocationService(t mockConstructorTestingTNewILocationService) *ILocationService {
	mock := &ILocationService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}