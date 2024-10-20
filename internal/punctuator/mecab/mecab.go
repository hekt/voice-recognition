package mecab

import (
	"fmt"
	"strings"

	mymecab "github.com/hekt/voice-recognition/internal/interfaces/mecab"
	"github.com/hekt/voice-recognition/internal/punctuator"
	"github.com/shogo82148/go-mecab"
)

type node struct {
	surface  string
	part     string
	partType string
	form     string
}

var (
	period = "。"
	comma  = "、"
	space  = " "
)

var _ punctuator.PunctuatorInterface = &MecabPunctuator{}

type MecabPunctuator struct {
	mecab   mymecab.MeCab
	builder strings.Builder
}

func NewMecabPunctuator(mecab mymecab.MeCab) (*MecabPunctuator, error) {
	return &MecabPunctuator{mecab: mecab}, nil
}

// Punctuate は与えられた文字列を句読点で区切り、余分なスペースを削除した文字列を返す。
// sentence はスペース区切りの文字列。空白の位置が句読点の挿入位置候補となる。
func (p *MecabPunctuator) Punctuate(sentence string) (string, error) {
	mecabNode, err := p.mecab.ParseToNode(sentence)
	if err != nil {
		return "", fmt.Errorf("failed to parse sentence: %w", err)
	}

	p.builder.Reset()
	prevNode := node{}
	// 最初のノードは surface を持たない BeginOfSentence node なので無視できる。
	for n := mecabNode.Next(); n.Stat() == mecab.NormalNode || n.Stat() == mecab.UnknownNode; n = n.Next() {
		// 各ノードの先頭にある空白に対して以下のいずれかの処理をする。
		// - そのまま空白として挿入する
		// - 空白を削除する
		// - 空白の代わりに句点を挿入する
		// - 空白の代わりに読点を挿入する
		node, hasSpace := parseNode(&n)
		if hasSpace {
			p.builder.WriteString(getPunctuation(prevNode, node))
		}

		prevNode = node
		p.builder.WriteString(node.surface)
	}

	return p.builder.String(), nil
}

func parseNode(n *mecab.Node) (node, bool) {
	// n.Length は先頭の空白を含まない文字数。n.RLength は先頭の空白を含む文字数。
	// これらが異なる場合は先頭に空白があるということ。
	hasSpace := n.Length() < n.RLength()

	fs := strings.Split(n.Feature(), ",")
	return node{
		surface:  n.Surface(),
		part:     fs[0],
		partType: fs[1],
		form:     fs[5],
	}, hasSpace
}

func getPunctuation(prev, next node) (p string) {
	if prev.partType == "終助詞" {
		if next.part != "助詞" {
			return period
		}
		if next.partType == "終助詞" {
			return
		}
	}

	// 違和感があった場合を除外している。
	if prev.part == "動詞" || prev.part == "形容詞" || prev.part == "助詞" {
		if prev.form == "基本形" &&
			prev.partType != "非自立" &&
			next.partType != "非自立" &&
			next.part != "名詞" &&
			next.part != "助詞" &&
			next.part != "助動詞" {
			return period
		}
	}

	if next.part == "フィラー" || prev.part == "フィラー" {
		return comma
	}

	if prev.part == "感動詞" {
		return comma
	}

	// 接続されていると違和感がある場合があるので、スペースを入れる
	// e.g. まあ年一回二回ぐらいがちょうどいいのでは[ ]ちょっと２年たったって思わなかったでしょ
	if prev.partType == "係助詞" && next.partType == "助詞類接続" {
		return space
	}
	// e.g. コラボウィークっていうのをやってて[ ]その中の何か三日目か四日目
	if next.part == "連体詞" {
		// 違和感がない気がするのでとりあえず読点。違和感がある箇所があればスペースにする。
		return comma
	}

	switch prev.part {
	case "名詞":
		if next.part == "動詞" {
			return
		}
		if next.part == "名詞" {
			return
		}
	case "副詞":
		if next.part == "名詞" {
			return
		}
	case "動詞":
		if next.part == "動詞" {
			return
		}
	case "形容詞", "接頭詞", "記号":
		return
	}

	switch next.part {
	case "名詞":
		if next.partType == "接尾" {
			return
		}
	case "動詞", "助詞", "助動詞", "記号":
		return
	}

	switch prev.partType {
	case "非自立", "助詞類接続", "連体化", "係助詞":
		return
	case "接続助詞":
		if next.part != "形容詞" && next.part != "副詞" && next.part != "名詞" {
			return
		}
	}

	switch next.partType {
	case "非自立":
		return
	}

	return space
}
