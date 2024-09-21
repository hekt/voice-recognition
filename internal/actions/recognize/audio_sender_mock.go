// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package recognize

import (
	"context"
	"sync"
)

// Ensure, that AudioSenderMock does implement AudioSender.
// If this is not the case, regenerate this file with moq.
var _ AudioSender = &AudioSenderMock{}

// AudioSenderMock is a mock implementation of AudioSender.
//
//	func TestSomethingThatUsesAudioSender(t *testing.T) {
//
//		// make and configure a mocked AudioSender
//		mockedAudioSender := &AudioSenderMock{
//			StartFunc: func(ctx context.Context) error {
//				panic("mock out the Start method")
//			},
//		}
//
//		// use mockedAudioSender in code that requires AudioSender
//		// and then make assertions.
//
//	}
type AudioSenderMock struct {
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
func (mock *AudioSenderMock) Start(ctx context.Context) error {
	if mock.StartFunc == nil {
		panic("AudioSenderMock.StartFunc: method is nil but AudioSender.Start was just called")
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
//	len(mockedAudioSender.StartCalls())
func (mock *AudioSenderMock) StartCalls() []struct {
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
