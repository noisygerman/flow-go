// Code generated by mockery v1.0.0. DO NOT EDIT.

package mock

import (
	flow "github.com/dapperlabs/flow-go/model/flow"
	message "github.com/dapperlabs/flow-go/network/gossip/libp2p/message"
	middleware "github.com/dapperlabs/flow-go/network/gossip/libp2p/middleware"

	mock "github.com/stretchr/testify/mock"

	time "time"
)

// Middleware is an autogenerated mock type for the Middleware type
type Middleware struct {
	mock.Mock
}

// Ping provides a mock function with given fields: targetID
func (_m *Middleware) Ping(targetID flow.Identifier) (time.Duration, error) {
	ret := _m.Called(targetID)

	var r0 time.Duration
	if rf, ok := ret.Get(0).(func(flow.Identifier) time.Duration); ok {
		r0 = rf(targetID)
	} else {
		r0 = ret.Get(0).(time.Duration)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(flow.Identifier) error); ok {
		r1 = rf(targetID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Send provides a mock function with given fields: channelID, msg, targetIDs
func (_m *Middleware) Send(channelID uint8, msg *message.Message, targetIDs ...flow.Identifier) error {
	_va := make([]interface{}, len(targetIDs))
	for _i := range targetIDs {
		_va[_i] = targetIDs[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, channelID, msg)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 error
	if rf, ok := ret.Get(0).(func(uint8, *message.Message, ...flow.Identifier) error); ok {
		r0 = rf(channelID, msg, targetIDs...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Start provides a mock function with given fields: overlay
func (_m *Middleware) Start(overlay middleware.Overlay) error {
	ret := _m.Called(overlay)

	var r0 error
	if rf, ok := ret.Get(0).(func(middleware.Overlay) error); ok {
		r0 = rf(overlay)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Stop provides a mock function with given fields:
func (_m *Middleware) Stop() {
	_m.Called()
}

// Subscribe provides a mock function with given fields: channelID
func (_m *Middleware) Subscribe(channelID uint8) error {
	ret := _m.Called(channelID)

	var r0 error
	if rf, ok := ret.Get(0).(func(uint8) error); ok {
		r0 = rf(channelID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateAllowlist provides a mock function with given fields:
func (_m *Middleware) UpdateAllowList() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
