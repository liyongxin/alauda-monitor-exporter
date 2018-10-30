FROM golang:1.10 as builder

WORKDIR $GOPATH/src/alauda-monitor-exporter
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o alauda_monitor_exporter .
RUN go install -v


FROM index.alauda.cn/alaudaorg/alaudabase:alpine-supervisor-migrate-1

WORKDIR /

COPY --from=builder /go/src/alauda-monitor-exporter .

RUN chmod +x /alauda_monitor_exporter
EXPOSE 8888

CMD ["/alauda_monitor_exporter"]
