# voice-recognition

CLI で文字起こしをするツール。

Google Cloud Speech-to-Text API を使っているためそのコストがかかる（執筆時点では $0.96/hour）

とりあえず Mac で動かすための手順を書いておく。

## 必要なもの

- gcloud CLI
- GStreamer
- BlackHole 2ch
- mecab 0.996
- (Google Cloud Speech-to-Text API)
  - `speech.recognizers.create`
  - `speech.recognizers.delete`
  - `speech.recognizers.list`
  - `speech.recognizers.recognize`

## 手順

### MeCab のインストール

vosk で使うために MeCab をインストールする。依存の関係で vosk を使わなくても必須。

```shell
brew install mecab mecab-ipadic

export CGO_LDLAGS="$(mecab-config --libs)"
export CGO_CFLAGS="-I$(mecab-config --incdir)"
```

### Mac のオーディオ設定

1. Audio MIDI 設定から BalckHole 2ch の出力 Format を 16000Hz に変更する
    - gstreamer での再サンプル処理を省いて高速化するため
2. 音声の出力先に BlackHole 2ch を指定する
    - これだけだと音を聴けないので、必要に応じて Audio MIDI 設定で BlackHole 2ch を含む複数出力装置を作成し、それを出力先にする

### gcloud を使う場合

#### gcloud の設定

```shell
gcloud auth application-default login
```

#### Google Cloud 上に Recognizer を作成

```shell
go run cmd/main.go recognizer-create \
    --project <project> \
    --recognizer <recognizerName> \
    --model long \
    --language ja-jp
```

Google Cloud 上に `recognizerName` という名前の Recognizer が作成される。 `recognizerName` はなんでもいいが、実行時に同じものを指定する必要がある。

#### recognize の実行

GStreamer で音声を取得して、それを Google Cloud Speech-to-Text API に投げる。

中間応答は標準出力、確定した結果はファイルに出力されるので、ファイルの出力を `tail -f` で見つつ標準出力で中間応答を見るといい。

```shell
gst-launch-1.0 -q osxaudiosrc device=<deviceNo> \
        ! audio/x-raw,format=S16LE,channels=1,rate=16000 \
        ! queue \
        ! fdsink fd=1 sync=false blocksize=4096 \
    | go run cmd/main.go recognize \
        --project <project> \
        --recognizer <recognizerName> \
        --buffersize 4096 \
        --output output.txt
```

- ctrl-c で終了する
- `deviceNo` は `say -a '?'` を実行すると得られる BlackHole 2ch の番号
- サンプリングレートが合わない場合は audoresample を追加する必要があるが、blackhole 2ch で合わせていれば不要
  - format, channels の調整は osxaudiosrc でやってくれるっぽい（audioconvert が必要な場合もあるかも）
- blocksize, buffersize は同じ値にする
  - [公式のベストプラクティス](https://cloud.google.com/speech-to-text/docs/best-practices-provide-speech-data?hl=ja#:~:text=100%20%E3%83%9F%E3%83%AA%E7%A7%92%E3%83%95%E3%83%AC%E3%83%BC%E3%83%A0%E3%82%B5%E3%82%A4%E3%82%BA%E3%82%92%E3%81%8A%E3%81%99%E3%81%99%E3%82%81%E3%81%97%E3%81%BE%E3%81%99%E3%80%82)にしたがって 100ms に近いフレームサイズになる数値にする
  - 16bit * 16000Hz * 0.1s = 3200byte なので近いところで 4096byte (128ms)

### Vosk を使う場合

```shell
cd path/to/repository
cd lib
curl -O https://alphacephei.com/vosk/models/vosk-model-small-ja-0.4.zip
unzip vosk-model-small-ja-0.4.zip
curl -O https://github.com/alphacep/vosk-api/releases/download/v0.3.42/vosk-osx-0.3.42.zip
unzip vosk-osx-0.3.42.zip
cd ../

export VOSK_PATH=/path/to/repository/lib/vosk-osx-0.3.42
export DYLD_LIBRARY_PATH=$VOSK_PATH
export CGO_CPPFLAGS="-I $VOSK_PATH $CGO_CPPFLAGS"
export CGO_LDFLAGS="-L $VOSK_PATH $CGO_LDFLAGS"

gst-launch-1.0 -q osxaudiosrc device=<deviceNo> \
        ! audio/x-raw,format=S16LE,channels=1,rate=16000 \
        ! queue \
        ! fdsink fd=1 sync=false blocksize=4096 \
    | go run cmd/main.go recognize-vosk \
        --model lib/vosk-model-small-ja-0.4 \
        --buffersize 4096 \
        --output output.txt
```

- `lib` でやっているが path の対応さえとれていれば別にどこでもよい
- dylib の読み込みは macOS にブロックされるので明示的に許可する必要がある
