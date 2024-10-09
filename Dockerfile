FROM golang:1.23 AS builder

WORKDIR /src

ENV GOCACHE=/go-cache
COPY go.mod .
RUN go mod download

COPY /cmd ./cmd/
COPY /ui ./ui/

RUN --mount=type=cache,target="/go-cache" go test -v ./...

ENV CGO_ENABLED=0
ENV GOOS=linux
RUN --mount=type=cache,target="/go-cache" go build -o /downtown ./cmd/downtown

FROM alpine

COPY --from=builder /downtown .

EXPOSE 4000
ENTRYPOINT ["/downtown"]
