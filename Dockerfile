# Build-Stage
FROM golang:alpine3.21 AS builder
RUN addgroup -g 1001 app && \
    adduser --system -u 1001 -G app app


WORKDIR /build


ARG APP_NAME="cloud-pdflatex"
ARG BUILD_COMMIT=""
ARG BUILD_TIME="sss"
ARG BUILD_VERSION="4567"
ARG BUILD_ID="__"
ARG GOOS="linux"
ARG GOARCH="arm64"




COPY . .
RUN go generate && \
    go mod download
RUN CGO_ENABLED=0 GOOS=$GOOS GOARCH=$GOARCH go build \
       -ldflags="-X main.buildVersion=${BUILD_VERSION}  -X main.appName=${APP_NAME}  -X main.buildCommit=${BUILD_COMMIT} -X main.buildTime=${BUILD_TIME} -w -s" \
        -o app .


FROM leplusorg/latex:main AS latex

# Set the working directory
WORKDIR /app
ENV CONF_LATEX_COMMAND="pdflatexmk"
COPY --from=builder /build/app /app/server

# Leaving user latex from parent image

# Define entrypoint and command
ENTRYPOINT ["/app/server"]