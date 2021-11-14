FROM node:12.18.4-alpine AS frontend_builder

COPY ./frontend /app/frontend
WORKDIR /app/frontend

ARG NPM_BUILD_MODE="development"
RUN npm install
RUN npm run build -- --mode "$NPM_BUILD_MODE"

FROM golang:1.16.7 AS backend_builder

COPY ./backend /app/backend

WORKDIR /app/backend

ARG GO_BUILD_TAGS="dev"
RUN go build --tags "$GO_BUILD_TAGS" -o /app/main main.go

FROM golang:1.16.7 AS litestream_builder

ARG litestream_version="0.3.6"

RUN set -x && \
    apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y \
      ca-certificates \
      git

RUN set -x && \
    git clone https://github.com/benbjohnson/litestream.git && \
    cd litestream && \
    git checkout "v${litestream_version}"

RUN set -x && \
    cd litestream && \
    go install ./cmd/litestream && \
    echo "litestream installed to ${GOPATH}/bin/litestream"

FROM golang:1.16.7

COPY --from=frontend_builder /app/frontend/dist /app/frontend/dist
COPY --from=backend_builder /app/main /app/main
COPY --from=litestream_builder /go/bin/litestream /app/litestream
COPY ./litestream.yml /etc/litestream.yml
COPY ./docker_entrypoint /app/docker_entrypoint

WORKDIR /app

ENTRYPOINT ["/app/docker_entrypoint"]
