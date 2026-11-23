FROM golang:1.27

WORKDIR /app

RUN apt-get -y update && apt-get install -y --no-install-recommends ffmpeg

COPY . .

RUN go mod download
RUN go build -o /server ./cmd/server

EXPOSE 8080

CMD ["/server"]
