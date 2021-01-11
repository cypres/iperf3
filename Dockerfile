FROM golang:alpine AS builder
COPY health .
RUN GOOS=linux GOARCH=amd64 go build -ldflags '-w -s' -o iperf-probe main.go


FROM alpine:latest AS runner

# build intial apk binary cache and install iperf3
RUN apk add --no-cache iperf3 \
    && adduser -S iperf

COPY --from=builder /go/iperf-probe /usr/local/bin/iperf-probe

# USER iperf

# Expose the default iperf3 server ports
EXPOSE 5201/tcp 5201/udp

# entrypoint allows you to pass your arguments to the container at runtime
# very similar to a binary you would run. For example, in the following
# docker run -it <IMAGE> --help' is like running 'iperf --help'
ENTRYPOINT ["iperf3"]

# iperf3 -s = run in Server mode
CMD ["-s"]
