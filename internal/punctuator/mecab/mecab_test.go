package mecab

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	mymecab "github.com/hekt/voice-recognition/internal/interfaces/mecab"
	"github.com/shogo82148/go-mecab"
)

func TestNewMecabPunctuator(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		got, err := NewMecabPunctuator(&mymecab.MeCabMock{})
		if err != nil {
			t.Errorf("NewMecabPunctuator() error = %v, want nil", err)
		}
		if got == nil {
			t.Error("NewMecabPunctuator() = nil, want not nil")
		}
	})
}

func TestMecabPunctuator_Punctuate(t *testing.T) {
	type args struct {
		sentence string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				sentence: "あのー そう だ わ ちょっと アレ の 話 も し とく か あれ の 話 も ある ま 石 食っ た 話 で も し ましょう か はい",
			},
			want: "あのー、そうだわ。ちょっとアレの話もしとくか あれの話もある。ま、石食った 話でもしましょうか はい",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := mecab.New(map[string]string{})
			if err != nil {
				t.Fatal(err)
			}
			defer m.Destroy()
			if _, err := m.Parse(""); err != nil {
				t.Fatal(err)
			}

			p := &MecabPunctuator{
				mecab:   m,
				builder: strings.Builder{},
			}
			got, err := p.Punctuate(tt.args.sentence)
			if (err != nil) != tt.wantErr {
				t.Errorf("MecabPunctuator.Punctuate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("MecabPunctuator.Punctuate() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func Test_getPunctuation(t *testing.T) {
	type args struct {
		prev node
		next node
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "空白: *-*",
			args: args{
				prev: node{},
				next: node{},
			},
			want: " ",
		},
		{
			name: "句点: 終助詞-助詞以外",
			args: args{
				prev: node{partType: "終助詞"},
				next: node{},
			},
			want: "。",
		},
		{
			name: "削除: 終助詞-助詞",
			args: args{
				prev: node{partType: "終助詞"},
				next: node{part: "助詞"},
			},
			want: "", // next.part == "助詞" のルールで削除される
		},
		{
			name: "削除: 終助詞-助詞_終助詞",
			args: args{
				prev: node{partType: "終助詞"},
				next: node{part: "助詞", partType: "終助詞"},
			},
			want: "",
		},
		{
			name: "句点: 動詞_基本形-動詞",
			args: args{
				prev: node{part: "動詞", form: "基本形"},
				next: node{part: "動詞"},
			},
			want: "。",
		},
		{
			name: "句点: 形容詞_基本形-動詞",
			args: args{
				prev: node{part: "形容詞", form: "基本形"},
				next: node{part: "動詞"},
			},
			want: "。",
		},
		{
			name: "句点: 助詞_基本形-動詞",
			args: args{
				prev: node{part: "助詞", form: "基本形"},
				next: node{part: "動詞"},
			},
			want: "。",
		},
		{
			name: "削除: 動詞_非自立_基本形-動詞",
			args: args{
				prev: node{part: "動詞", form: "基本形", partType: "非自立"},
				next: node{part: "動詞"},
			},
			want: "", // prev.partType == "非自立" のルールで削除される
		},
		{
			name: "削除: 動詞_基本形-非自立",
			args: args{
				prev: node{part: "動詞", form: "基本形"},
				next: node{partType: "非自立"},
			},
			want: "", // next.partType == "非自立" のルールで削除される
		},
		{
			name: "削除: 動詞_基本形-名詞",
			args: args{
				prev: node{part: "動詞", form: "基本形"},
				next: node{part: "名詞"},
			},
			want: " ",
		},
		{
			name: "削除: 動詞_基本形-助詞",
			args: args{
				prev: node{part: "動詞", form: "基本形"},
				next: node{part: "助詞"},
			},
			want: "", // next.part == "助詞" のルールで削除される
		},
		{
			name: "削除: 動詞_基本形-助動詞",
			args: args{
				prev: node{part: "動詞", form: "基本形"},
				next: node{part: "助動詞"},
			},
			want: "", // next.type == "助動詞" のルールで削除される
		},
		{
			name: "読点: フィラー-*",
			args: args{
				prev: node{part: "フィラー"},
				next: node{},
			},
			want: "、",
		},
		{
			name: "読点: *-フィラー",
			args: args{
				prev: node{},
				next: node{part: "フィラー"},
			},
			want: "、",
		},
		{
			name: "読点: 感動詞-*",
			args: args{
				prev: node{part: "感動詞"},
				next: node{},
			},
			want: "、",
		},
		{
			name: "空白: 係助詞-助詞類接続",
			args: args{
				prev: node{partType: "係助詞"},
				next: node{partType: "助詞類接続"},
			},
			want: " ",
		},
		{
			name: "読点: *-連体詞",
			args: args{
				prev: node{},
				next: node{part: "連体詞"},
			},
			want: "、",
		},
		{
			name: "削除: 名詞-動詞",
			args: args{
				prev: node{part: "名詞"},
				next: node{part: "動詞"},
			},
			want: "",
		},
		{
			name: "削除: 名詞-名詞",
			args: args{
				prev: node{part: "名詞"},
				next: node{part: "名詞"},
			},
			want: "",
		},
		{
			name: "削除: 副詞-名詞",
			args: args{
				prev: node{part: "副詞"},
				next: node{part: "名詞"},
			},
			want: "",
		},
		{
			name: "削除: 動詞-動詞",
			args: args{
				prev: node{part: "動詞"},
				next: node{part: "動詞"},
			},
			want: "",
		},
		{
			name: "削除: 形容詞-*",
			args: args{
				prev: node{part: "形容詞"},
				next: node{},
			},
			want: "",
		},
		{
			name: "削除: 接頭詞-*",
			args: args{
				prev: node{part: "接頭詞"},
				next: node{},
			},
			want: "",
		},
		{
			name: "削除: 記号-*",
			args: args{
				prev: node{part: "記号"},
				next: node{},
			},
			want: "",
		},
		{
			name: "削除: *-名詞_接尾",
			args: args{
				prev: node{},
				next: node{part: "名詞", partType: "接尾"},
			},
			want: "",
		},
		{
			name: "削除: *-動詞",
			args: args{
				prev: node{},
				next: node{part: "動詞"},
			},
			want: "",
		},
		{
			name: "削除: *-助詞",
			args: args{
				prev: node{},
				next: node{part: "助詞"},
			},
			want: "",
		},
		{
			name: "削除: *-助動詞",
			args: args{
				prev: node{},
				next: node{part: "助動詞"},
			},
			want: "",
		},
		{
			name: "削除: *-記号",
			args: args{
				prev: node{},
				next: node{part: "記号"},
			},
			want: "",
		},
		{
			name: "削除: 非自立-*",
			args: args{
				prev: node{partType: "非自立"},
				next: node{},
			},
			want: "",
		},
		{
			name: "削除: 助詞類接続-*",
			args: args{
				prev: node{partType: "助詞類接続"},
				next: node{},
			},
			want: "",
		},
		{
			name: "削除: 連体化-*",
			args: args{
				prev: node{partType: "連体化"},
				next: node{},
			},
			want: "",
		},
		{
			name: "削除: 係助詞-*",
			args: args{
				prev: node{partType: "係助詞"},
				next: node{},
			},
			want: "",
		},
		{
			name: "削除: 接続助詞-*",
			args: args{
				prev: node{partType: "接続助詞"},
				next: node{},
			},
			want: "",
		},
		{
			name: "空白: 接続助詞-形容詞",
			args: args{
				prev: node{partType: "接続助詞"},
				next: node{part: "形容詞"},
			},
			want: " ",
		},
		{
			name: "空白: 接続助詞-副詞",
			args: args{
				prev: node{partType: "接続助詞"},
				next: node{part: "副詞"},
			},
			want: " ",
		},
		{
			name: "空白: 接続助詞-名詞",
			args: args{
				prev: node{partType: "接続助詞"},
				next: node{part: "名詞"},
			},
			want: " ",
		},
		{
			name: "削除: *-非自立",
			args: args{
				prev: node{},
				next: node{partType: "非自立"},
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotChar := getPunctuation(tt.args.prev, tt.args.next); gotChar != tt.want {
				t.Errorf("getPunctuation() = %v, want %v", gotChar, tt.want)
			}
		})
	}
}
