// Code generated by mockery v1.1.2. DO NOT EDIT.

package main

import mock "github.com/stretchr/testify/mock"

// MockConnPool is an autogenerated mock type for the ConnPool type
type MockConnPool struct {
	mock.Mock
}

// Add provides a mock function with given fields: ce
func (_m *MockConnPool) Add(ce *ConnEntry) error {
	ret := _m.Called(ce)

	var r0 error
	if rf, ok := ret.Get(0).(func(*ConnEntry) error); ok {
		r0 = rf(ce)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// AddUpstream provides a mock function with given fields: r
func (_m *MockConnPool) AddUpstream(r *Upstream) {
	_m.Called(r)
}

// CloseConnection provides a mock function with given fields: ce
func (_m *MockConnPool) CloseConnection(ce *ConnEntry) {
	_m.Called(ce)
}

// Get provides a mock function with given fields:
func (_m *MockConnPool) Get() (*ConnEntry, Upstream) {
	ret := _m.Called()

	var r0 *ConnEntry
	if rf, ok := ret.Get(0).(func() *ConnEntry); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*ConnEntry)
		}
	}

	var r1 Upstream
	if rf, ok := ret.Get(1).(func() Upstream); ok {
		r1 = rf()
	} else {
		r1 = ret.Get(1).(Upstream)
	}

	return r0, r1
}

// NewConnection provides a mock function with given fields: upstream, dialFunc
func (_m *MockConnPool) NewConnection(upstream Upstream, dialFunc DialFunc) (*ConnEntry, error) {
	ret := _m.Called(upstream, dialFunc)

	var r0 *ConnEntry
	if rf, ok := ret.Get(0).(func(Upstream, DialFunc) *ConnEntry); ok {
		r0 = rf(upstream, dialFunc)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*ConnEntry)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(Upstream, DialFunc) error); ok {
		r1 = rf(upstream, dialFunc)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Size provides a mock function with given fields:
func (_m *MockConnPool) Size() int {
	ret := _m.Called()

	var r0 int
	if rf, ok := ret.Get(0).(func() int); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int)
	}

	return r0
}