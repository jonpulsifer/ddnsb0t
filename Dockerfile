FROM golang:alpine AS build
# hadolint ignore=DL3018
RUN apk update && apk add --no-cache git
WORKDIR /go/src/github.com/jonpulsifer/ddnsb0t
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -installsuffix cgo -ldflags '-w -s' -o /go/bin/ddnsb0t

# hadolint ignore=DL3007
FROM gcr.io/distroless/static:latest
USER 65534:65534
COPY --from=build /go/bin/ddnsb0t /ddnsb0t
ENTRYPOINT ["/ddnsb0t"]
CMD ["--help"]
