// Code generated by mockery v2.52.2. DO NOT EDIT.

package mocks

import (
	context "context"

	events "github.com/aws/aws-lambda-go/events"
	mock "github.com/stretchr/testify/mock"
)

// IBlockService is an autogenerated mock type for the IBlockService type
type IBlockService struct {
	mock.Mock
}

// ProcessBlock provides a mock function with given fields: ctx, event
func (_m *IBlockService) ProcessBlock(ctx context.Context, event events.SQSEvent) error {
	ret := _m.Called(ctx, event)

	if len(ret) == 0 {
		panic("no return value specified for ProcessBlock")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, events.SQSEvent) error); ok {
		r0 = rf(ctx, event)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewIBlockService creates a new instance of IBlockService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewIBlockService(t interface {
	mock.TestingT
	Cleanup(func())
}) *IBlockService {
	mock := &IBlockService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
