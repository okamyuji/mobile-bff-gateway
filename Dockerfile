FROM golang:1.24-alpine AS build

WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal
RUN go build -o /out/gateway ./cmd/gateway && go build -o /out/mockservice ./cmd/mockservice

FROM alpine:3.21

COPY --from=build /out/gateway /usr/local/bin/gateway
COPY --from=build /out/mockservice /usr/local/bin/mockservice
ENTRYPOINT ["/usr/local/bin/gateway"]
