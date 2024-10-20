// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package punctuator

import (
	"sync"
)

// Ensure, that PunctuatorInterfaceMock does implement PunctuatorInterface.
// If this is not the case, regenerate this file with moq.
var _ PunctuatorInterface = &PunctuatorInterfaceMock{}

// PunctuatorInterfaceMock is a mock implementation of PunctuatorInterface.
//
//	func TestSomethingThatUsesPunctuatorInterface(t *testing.T) {
//
//		// make and configure a mocked PunctuatorInterface
//		mockedPunctuatorInterface := &PunctuatorInterfaceMock{
//			PunctuateFunc: func(sentence string) (string, error) {
//				panic("mock out the Punctuate method")
//			},
//		}
//
//		// use mockedPunctuatorInterface in code that requires PunctuatorInterface
//		// and then make assertions.
//
//	}
type PunctuatorInterfaceMock struct {
	// PunctuateFunc mocks the Punctuate method.
	PunctuateFunc func(sentence string) (string, error)

	// calls tracks calls to the methods.
	calls struct {
		// Punctuate holds details about calls to the Punctuate method.
		Punctuate []struct {
			// Sentence is the sentence argument value.
			Sentence string
		}
	}
	lockPunctuate sync.RWMutex
}

// Punctuate calls PunctuateFunc.
func (mock *PunctuatorInterfaceMock) Punctuate(sentence string) (string, error) {
	if mock.PunctuateFunc == nil {
		panic("PunctuatorInterfaceMock.PunctuateFunc: method is nil but PunctuatorInterface.Punctuate was just called")
	}
	callInfo := struct {
		Sentence string
	}{
		Sentence: sentence,
	}
	mock.lockPunctuate.Lock()
	mock.calls.Punctuate = append(mock.calls.Punctuate, callInfo)
	mock.lockPunctuate.Unlock()
	return mock.PunctuateFunc(sentence)
}

// PunctuateCalls gets all the calls that were made to Punctuate.
// Check the length with:
//
//	len(mockedPunctuatorInterface.PunctuateCalls())
func (mock *PunctuatorInterfaceMock) PunctuateCalls() []struct {
	Sentence string
} {
	var calls []struct {
		Sentence string
	}
	mock.lockPunctuate.RLock()
	calls = mock.calls.Punctuate
	mock.lockPunctuate.RUnlock()
	return calls
}
