// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package recognizer

import (
	"context"
	"sync"
)

// Ensure, that ResultWriterInterfaceMock does implement ResultWriterInterface.
// If this is not the case, regenerate this file with moq.
var _ ResultWriterInterface = &ResultWriterInterfaceMock{}

// ResultWriterInterfaceMock is a mock implementation of ResultWriterInterface.
//
//	func TestSomethingThatUsesResultWriterInterface(t *testing.T) {
//
//		// make and configure a mocked ResultWriterInterface
//		mockedResultWriterInterface := &ResultWriterInterfaceMock{
//			StartFunc: func(ctx context.Context) error {
//				panic("mock out the Start method")
//			},
//		}
//
//		// use mockedResultWriterInterface in code that requires ResultWriterInterface
//		// and then make assertions.
//
//	}
type ResultWriterInterfaceMock struct {
	// StartFunc mocks the Start method.
	StartFunc func(ctx context.Context) error

	// calls tracks calls to the methods.
	calls struct {
		// Start holds details about calls to the Start method.
		Start []struct {
			// Ctx is the ctx argument value.
			Ctx context.Context
		}
	}
	lockStart sync.RWMutex
}

// Start calls StartFunc.
func (mock *ResultWriterInterfaceMock) Start(ctx context.Context) error {
	if mock.StartFunc == nil {
		panic("ResultWriterInterfaceMock.StartFunc: method is nil but ResultWriterInterface.Start was just called")
	}
	callInfo := struct {
		Ctx context.Context
	}{
		Ctx: ctx,
	}
	mock.lockStart.Lock()
	mock.calls.Start = append(mock.calls.Start, callInfo)
	mock.lockStart.Unlock()
	return mock.StartFunc(ctx)
}

// StartCalls gets all the calls that were made to Start.
// Check the length with:
//
//	len(mockedResultWriterInterface.StartCalls())
func (mock *ResultWriterInterfaceMock) StartCalls() []struct {
	Ctx context.Context
} {
	var calls []struct {
		Ctx context.Context
	}
	mock.lockStart.RLock()
	calls = mock.calls.Start
	mock.lockStart.RUnlock()
	return calls
}
