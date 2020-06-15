// Code generated by mockery v1.0.0. DO NOT EDIT.

package mock

import (
	flow "github.com/dapperlabs/flow-go/model/flow"
	mock "github.com/stretchr/testify/mock"
)

// Blocks is an autogenerated mock type for the Blocks type
type Blocks struct {
	mock.Mock
}

// ByCollectionID provides a mock function with given fields: collID
func (_m *Blocks) ByCollectionID(collID flow.Identifier) (*flow.Block, error) {
	ret := _m.Called(collID)

	var r0 *flow.Block
	if rf, ok := ret.Get(0).(func(flow.Identifier) *flow.Block); ok {
		r0 = rf(collID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*flow.Block)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(flow.Identifier) error); ok {
		r1 = rf(collID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ByHeight provides a mock function with given fields: height
func (_m *Blocks) ByHeight(height uint64) (*flow.Block, error) {
	ret := _m.Called(height)

	var r0 *flow.Block
	if rf, ok := ret.Get(0).(func(uint64) *flow.Block); ok {
		r0 = rf(height)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*flow.Block)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(uint64) error); ok {
		r1 = rf(height)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ByID provides a mock function with given fields: blockID
func (_m *Blocks) ByID(blockID flow.Identifier) (*flow.Block, error) {
	ret := _m.Called(blockID)

	var r0 *flow.Block
	if rf, ok := ret.Get(0).(func(flow.Identifier) *flow.Block); ok {
		r0 = rf(blockID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*flow.Block)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(flow.Identifier) error); ok {
		r1 = rf(blockID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// IndexBlockForCollections provides a mock function with given fields: blockID, collIDs
func (_m *Blocks) IndexBlockForCollections(blockID flow.Identifier, collIDs []flow.Identifier) error {
	ret := _m.Called(blockID, collIDs)

	var r0 error
	if rf, ok := ret.Get(0).(func(flow.Identifier, []flow.Identifier) error); ok {
		r0 = rf(blockID, collIDs)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Store provides a mock function with given fields: block
func (_m *Blocks) Store(block *flow.Block) error {
	ret := _m.Called(block)

	var r0 error
	if rf, ok := ret.Get(0).(func(*flow.Block) error); ok {
		r0 = rf(block)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
