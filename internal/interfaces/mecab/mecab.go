package mecab

import "github.com/shogo82148/go-mecab"

//go:generate moq -rm -out mecab_mock.go . MeCab
type MeCab interface {
	ParseToNode(string) (mecab.Node, error)
}
