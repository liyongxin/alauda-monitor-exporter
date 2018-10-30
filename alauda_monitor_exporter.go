
package main

import (
	"net/http"
	_ "net/http/pprof"
	"math/rand"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"
	"time"
)


var openFileNums = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Name: "open_files_num",
		Help: "node open file nums",
	},
	[]string{"message", "name"},
)

// Counter metric for request count
var yxliHttpRequestCount = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "yxli_http_request_count",
		Help: "http request count",
	},
	[]string{"path"},
)

func init() {
	prometheus.MustRegister(version.NewCollector("alauda_monitor_exporter"))
	prometheus.MustRegister(yxliHttpRequestCount)
	prometheus.MustRegister(openFileNums)
}

func main() {
	var (
		listenAddress = kingpin.Flag("web.listen-address", "Address on which to expose metrics and web interface.").Default(":6666").String()
		metricsPath   = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
	)

	log.AddFlags(kingpin.CommandLine)
	kingpin.Version(version.Print("alauda_monitor_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	log.Infoln("Starting alauda monitor exporter", version.Info())
	log.Infoln("Build context", version.BuildContext())
	rd := rand.New(rand.NewSource(time.Now().UnixNano()))

	openFileNums.With(prometheus.Labels{"name":"k8s-furion", "message":"i saw the world"}).Set(rd.Float64())
	yxliHttpRequestCount.WithLabelValues("/furion/gp").Inc()

	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Alauda Monitor Exporter</title></head>
			<body>
			<h1>Alauda Monitor Exporter</h1>
			<p><a href="` + *metricsPath + `">Metrics</a></p>
			</body>
			</html>`))
	})

	log.Infoln("Listening on", *listenAddress)
	err := http.ListenAndServe(*listenAddress, nil)
	if err != nil {
		log.Fatal(err)
	}
}
