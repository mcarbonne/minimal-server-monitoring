# Build stage
FROM golang:1.21-bookworm as buildstage

WORKDIR /src
COPY . /src/.
RUN make build

FROM alpine:3.19.1 as runtime

RUN apk add libc6-compat
COPY --from=buildstage --chmod=755 /src/minimal-server-monitoring /app/.
COPY docker_config.json /app/config.json

WORKDIR /app

CMD ["/app/minimal-server-monitoring", "config.json"]
