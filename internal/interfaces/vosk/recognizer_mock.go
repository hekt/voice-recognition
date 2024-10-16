// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package vosk

import (
	"sync"
)

// Ensure, that VoskRecognizerMock does implement VoskRecognizer.
// If this is not the case, regenerate this file with moq.
var _ VoskRecognizer = &VoskRecognizerMock{}

// VoskRecognizerMock is a mock implementation of VoskRecognizer.
//
//	func TestSomethingThatUsesVoskRecognizer(t *testing.T) {
//
//		// make and configure a mocked VoskRecognizer
//		mockedVoskRecognizer := &VoskRecognizerMock{
//			AcceptWaveformFunc: func(bytes []byte) int {
//				panic("mock out the AcceptWaveform method")
//			},
//			FinalResultFunc: func() []byte {
//				panic("mock out the FinalResult method")
//			},
//			PartialResultFunc: func() []byte {
//				panic("mock out the PartialResult method")
//			},
//			ResultFunc: func() []byte {
//				panic("mock out the Result method")
//			},
//		}
//
//		// use mockedVoskRecognizer in code that requires VoskRecognizer
//		// and then make assertions.
//
//	}
type VoskRecognizerMock struct {
	// AcceptWaveformFunc mocks the AcceptWaveform method.
	AcceptWaveformFunc func(bytes []byte) int

	// FinalResultFunc mocks the FinalResult method.
	FinalResultFunc func() []byte

	// PartialResultFunc mocks the PartialResult method.
	PartialResultFunc func() []byte

	// ResultFunc mocks the Result method.
	ResultFunc func() []byte

	// calls tracks calls to the methods.
	calls struct {
		// AcceptWaveform holds details about calls to the AcceptWaveform method.
		AcceptWaveform []struct {
			// Bytes is the bytes argument value.
			Bytes []byte
		}
		// FinalResult holds details about calls to the FinalResult method.
		FinalResult []struct {
		}
		// PartialResult holds details about calls to the PartialResult method.
		PartialResult []struct {
		}
		// Result holds details about calls to the Result method.
		Result []struct {
		}
	}
	lockAcceptWaveform sync.RWMutex
	lockFinalResult    sync.RWMutex
	lockPartialResult  sync.RWMutex
	lockResult         sync.RWMutex
}

// AcceptWaveform calls AcceptWaveformFunc.
func (mock *VoskRecognizerMock) AcceptWaveform(bytes []byte) int {
	if mock.AcceptWaveformFunc == nil {
		panic("VoskRecognizerMock.AcceptWaveformFunc: method is nil but VoskRecognizer.AcceptWaveform was just called")
	}
	callInfo := struct {
		Bytes []byte
	}{
		Bytes: bytes,
	}
	mock.lockAcceptWaveform.Lock()
	mock.calls.AcceptWaveform = append(mock.calls.AcceptWaveform, callInfo)
	mock.lockAcceptWaveform.Unlock()
	return mock.AcceptWaveformFunc(bytes)
}

// AcceptWaveformCalls gets all the calls that were made to AcceptWaveform.
// Check the length with:
//
//	len(mockedVoskRecognizer.AcceptWaveformCalls())
func (mock *VoskRecognizerMock) AcceptWaveformCalls() []struct {
	Bytes []byte
} {
	var calls []struct {
		Bytes []byte
	}
	mock.lockAcceptWaveform.RLock()
	calls = mock.calls.AcceptWaveform
	mock.lockAcceptWaveform.RUnlock()
	return calls
}

// FinalResult calls FinalResultFunc.
func (mock *VoskRecognizerMock) FinalResult() []byte {
	if mock.FinalResultFunc == nil {
		panic("VoskRecognizerMock.FinalResultFunc: method is nil but VoskRecognizer.FinalResult was just called")
	}
	callInfo := struct {
	}{}
	mock.lockFinalResult.Lock()
	mock.calls.FinalResult = append(mock.calls.FinalResult, callInfo)
	mock.lockFinalResult.Unlock()
	return mock.FinalResultFunc()
}

// FinalResultCalls gets all the calls that were made to FinalResult.
// Check the length with:
//
//	len(mockedVoskRecognizer.FinalResultCalls())
func (mock *VoskRecognizerMock) FinalResultCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockFinalResult.RLock()
	calls = mock.calls.FinalResult
	mock.lockFinalResult.RUnlock()
	return calls
}

// PartialResult calls PartialResultFunc.
func (mock *VoskRecognizerMock) PartialResult() []byte {
	if mock.PartialResultFunc == nil {
		panic("VoskRecognizerMock.PartialResultFunc: method is nil but VoskRecognizer.PartialResult was just called")
	}
	callInfo := struct {
	}{}
	mock.lockPartialResult.Lock()
	mock.calls.PartialResult = append(mock.calls.PartialResult, callInfo)
	mock.lockPartialResult.Unlock()
	return mock.PartialResultFunc()
}

// PartialResultCalls gets all the calls that were made to PartialResult.
// Check the length with:
//
//	len(mockedVoskRecognizer.PartialResultCalls())
func (mock *VoskRecognizerMock) PartialResultCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockPartialResult.RLock()
	calls = mock.calls.PartialResult
	mock.lockPartialResult.RUnlock()
	return calls
}

// Result calls ResultFunc.
func (mock *VoskRecognizerMock) Result() []byte {
	if mock.ResultFunc == nil {
		panic("VoskRecognizerMock.ResultFunc: method is nil but VoskRecognizer.Result was just called")
	}
	callInfo := struct {
	}{}
	mock.lockResult.Lock()
	mock.calls.Result = append(mock.calls.Result, callInfo)
	mock.lockResult.Unlock()
	return mock.ResultFunc()
}

// ResultCalls gets all the calls that were made to Result.
// Check the length with:
//
//	len(mockedVoskRecognizer.ResultCalls())
func (mock *VoskRecognizerMock) ResultCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockResult.RLock()
	calls = mock.calls.Result
	mock.lockResult.RUnlock()
	return calls
}
