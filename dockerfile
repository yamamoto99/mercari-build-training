# syntax=docker/dockerfile:1

# alpineからgoのversionを指定
FROM golang:1.22-alpine

# ワーキングディレクトリを/appに指定
WORKDIR /app

# コンパイルするためにGCCを追加
# alpineはmusl libcを採用してるのでCのプログラムをビルドするためにmusl-devが必要
RUN apk add --no-cache gcc musl-dev

# ファイルやディレクトリを指定して/appにコピーする
COPY go/go.mod /app
COPY go/go.sum /app
COPY db/mercari.sqlite3 /app
COPY go/app/main.go /app
# imagesはフォルダで持ちたいのでフォルダを/app/imagesにコピー
COPY go/images/ /app/images

# 依存関係を整理
RUN go mod tidy

# グループを作成しtraineeを作成し作成したグループに入れる
RUN addgroup -S mercari && adduser -S trainee -G mercari

# ビルドした時にデフォルトだと/appに書き込み権限がないので所有者とグループを変更
RUN chown -R trainee:mercari /app

# コンテナ内の実行をtraineeで行うことを設定
USER trainee

# /appをビルド
RUN go build -o a.out /app

# a.outを実行
CMD ["./a.out"]
