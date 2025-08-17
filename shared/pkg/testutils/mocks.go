package testutils

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
)

// CreateTestContext creates a test context with timeout
func CreateTestContext() context.Context {
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	return ctx
}

// MockExpectation represents a mock expectation helper
type MockExpectation struct {
	Mock   *mock.Mock
	Method string
	Args   []interface{}
	Return []interface{}
	Times  int
}

// SetupMockExpectations sets up multiple mock expectations at once
func SetupMockExpectations(expectations []MockExpectation) {
	for _, exp := range expectations {
		call := exp.Mock.On(exp.Method, exp.Args...)
		if len(exp.Return) > 0 {
			call.Return(exp.Return...)
		}
		if exp.Times > 0 {
			if exp.Times == 1 {
				call.Once()
			} else {
				call.Times(exp.Times)
			}
		}
	}
}

// AssertMockExpectations asserts all mock expectations
func AssertMockExpectations(t mock.TestingT, mocks ...*mock.Mock) {
	for _, m := range mocks {
		m.AssertExpectations(t)
	}
}
