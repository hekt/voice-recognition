package punctuator

//go:generate moq -rm -out punctuator_mock.go . PunctuatorInterface
type PunctuatorInterface interface {
	Punctuate(sentence string) (string, error)
}
