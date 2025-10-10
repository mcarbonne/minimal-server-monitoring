# Build stage
FROM golang:1.25.1-bookworm AS buildstage

WORKDIR /src
COPY . /src/.
RUN make build

FROM alpine:3.22.2 AS runtime

RUN apk add libc6-compat
COPY --from=buildstage --chmod=755 /src/minimal-server-monitoring /app/.
COPY docker_config.yml /app/config.yml

WORKDIR /app

CMD ["/app/minimal-server-monitoring", "config.yml"]
