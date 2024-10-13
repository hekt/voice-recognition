package vosk

//go:generate moq -rm -out recognizer_mock.go . VoskRecognizer
type VoskRecognizer interface {
	AcceptWaveform([]byte) int
	PartialResult() []byte
	Result() []byte
	FinalResult() []byte
}
