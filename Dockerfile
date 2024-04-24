FROM golang:latest

WORKDIR /app/api
COPY ./api .

# OSのインストール済みのパッケージをバージョンアップし、必要なパッケージをインストール
RUN apt-get update && \
    apt-get install -y git gcc musl-dev && \
    rm -rf /var/lib/apt/lists/*


# デバッグ用のツール
RUN go install github.com/go-delve/delve/cmd/dlv@latest


