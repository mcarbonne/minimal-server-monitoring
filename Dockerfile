# Build stage
FROM golang:1.25.5-bookworm AS buildstage

WORKDIR /src
COPY . /src/.
RUN make build

FROM alpine:3.23.0 AS runtime

RUN apk add libc6-compat
COPY --from=buildstage --chmod=755 /src/minimal-server-monitoring /app/.
COPY docker_config.yml /app/config.yml

WORKDIR /app

CMD ["/app/minimal-server-monitoring", "config.yml"]
