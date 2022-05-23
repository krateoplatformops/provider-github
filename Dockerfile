# Build environment
# -----------------
FROM golang:1.18.0-bullseye as builder
LABEL stage=builder

ARG DEBIAN_FRONTEND=noninteractive

SHELL ["/bin/bash", "-o", "pipefail", "-c"]
# hadolint ignore=DL3008
RUN apt-get update && apt-get install -y ca-certificates openssl git tzdata && apt-get install -y --no-install-recommends xz-utils && \
  update-ca-certificates && \
  apt-get remove -y xz-utils && \
  rm -rf /var/lib/apt/lists/*

WORKDIR /src

COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

COPY cmd/ cmd/
COPY apis/ apis/
COPY pkg/ pkg/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o /bin/controller cmd/main.go && \
    strip /bin/controller

# Deployment environment
# ----------------------
FROM scratch

COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=builder /bin/controller /bin/controller

# Metadata params
ARG VERSION
ARG BUILD_DATE
ARG REPO_URL
ARG LAST_COMMIT
ARG PROJECT_NAME
ARG VENDOR

# Metadata
LABEL org.label-schema.build-date=$BUILD_DATE \
      org.label-schema.name=$PROJECT_NAME \
      org.label-schema.vcs-url=$REPO_URL \
      org.label-schema.vcs-ref=$LAST_COMMIT \
      org.label-schema.vendor=$VENDOR \
      org.label-schema.version=$VERSION \
      org.label-schema.docker.schema-version="1.0"

ARG METRICS_PORT
ARG HEALTHZ_PORT

EXPOSE ${METRICS_PORT}
EXPOSE ${HEALTHZ_PORT}

ENTRYPOINT ["/bin/controller"]