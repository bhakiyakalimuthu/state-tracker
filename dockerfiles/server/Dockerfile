FROM golang:1.19-alpine3.17 as builder
ARG VERSION
ARG APP_NAME
ENV LISTEN_PORT=6666
WORKDIR /build
COPY go.mod .
COPY go.sum .
RUN go mod download
ADD . .

RUN apk add --no-cache
RUN --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=0 go build -trimpath -ldflags "-s -X main._BuildVersion=${VERSION} -X main._AppName=${APP_NAME}" -v -o ${APP_NAME} ./cmd/server/main.go


FROM alpine:latest
ARG VERSION
ARG APP_NAME
ENV LISTEN_PORT=9090
WORKDIR /app
COPY --from=builder /build/${APP_NAME} /app/${APP_NAME}
RUN chmod +x /app/${APP_NAME}
EXPOSE ${LISTEN_PORT}
ENV APP=/app/${APP_NAME}
CMD $APP
