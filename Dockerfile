FROM golang:alpine AS builder

WORKDIR /build

# Make sure dependencies are available before building
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY server ./server
COPY store ./store
COPY spam ./spam
COPY VERSION .
COPY main.go .
RUN go build

FROM alpine

COPY --from=builder /build/mmocg /

EXPOSE 5000

ENTRYPOINT ["/mmocg"]
