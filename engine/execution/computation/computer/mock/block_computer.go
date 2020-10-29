// Code generated by mockery v1.0.0. DO NOT EDIT.

package mock

import (
	context "context"

	delta "github.com/onflow/flow-go/engine/execution/state/delta"
	entity "github.com/onflow/flow-go/module/mempool/entity"

	execution "github.com/onflow/flow-go/engine/execution"

	mock "github.com/stretchr/testify/mock"
)

// BlockComputer is an autogenerated mock type for the BlockComputer type
type BlockComputer struct {
	mock.Mock
}

// ExecuteBlock provides a mock function with given fields: _a0, _a1, _a2
func (_m *BlockComputer) ExecuteBlock(_a0 context.Context, _a1 *entity.ExecutableBlock, _a2 *delta.View) (*execution.ComputationResult, error) {
	ret := _m.Called(_a0, _a1, _a2)

	var r0 *execution.ComputationResult
	if rf, ok := ret.Get(0).(func(context.Context, *entity.ExecutableBlock, *delta.View) *execution.ComputationResult); ok {
		r0 = rf(_a0, _a1, _a2)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*execution.ComputationResult)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *entity.ExecutableBlock, *delta.View) error); ok {
		r1 = rf(_a0, _a1, _a2)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
