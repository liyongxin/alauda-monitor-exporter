
package main

import (
	"net/http"
	_ "net/http/pprof"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"
	"io/ioutil"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)


type GlobalServiceDiagnose struct{
	serviceName string
	diagnoseUrl string
	healthDesc  *prometheus.Desc
}

type diagnoseKVPair map[string]string
var componenetList = map[string] string{
	"furion": "http://k8s-furion:8080",
	"jakiro": "http://k8s-jakiro",
}

// Describe simply sends the two Descs in the struct to the channel.
func (g *GlobalServiceDiagnose) Describe(ch chan<- *prometheus.Desc) {
	ch <- g.healthDesc
}

func (g *GlobalServiceDiagnose) Collect(ch chan<- prometheus.Metric) {
	diag := g.healthCheck()
	for comp, res := range *diag {
		health , err := strconv.ParseFloat(res["health"], 64)
		if err != nil {
			panic(err)
		}
		ch <- prometheus.MustNewConstMetric(
			g.healthDesc,
			prometheus.GaugeValue,
			float64(health),
			comp,
			res["message"],
		)
	}

}

/*
{
	"furion": {"health": 1, "message": "everything is ok"},
	"jakiro": {"health": 0, "message": "everything is bad"}
}
*/
func (g *GlobalServiceDiagnose) healthCheck ()  *map[string]diagnoseKVPair{
	res := make(map[string]diagnoseKVPair)
	furionDiag := diagnoseKVPair{
		"health": "1",
		"message": "everything is ok",
	}
	jakiroDiag := diagnoseKVPair{
		"health": "1",
		"message": "everything is bad",
	}
	res["phoenix"] = furionDiag
	res["jakiro"] = jakiroDiag
	return &res
}

func HttpGet(globalName string)  *diagnoseKVPair {
	url := componenetList[globalName]
	resp, err := http.Get(url)
	defer resp.Body.Close()
	if err != nil {
		log.Errorln(err)
	}
	dm := make(diagnoseKVPair)
	bts, _:= ioutil.ReadAll(resp.Body)
	//fmt.Println(string(bytessa))
	json.Unmarshal(bts, &dm)
	return &dm
}

func NewServiceDiagnoser(name, url string) *GlobalServiceDiagnose{
	return &GlobalServiceDiagnose{
		serviceName: name,
		diagnoseUrl: url,
		healthDesc: prometheus.NewDesc(
			"global_service_diagnose",
			fmt.Sprintf("global service %s diagnose.",name),
			[]string{"global_service_name", "message"},
			prometheus.Labels{"global_service_name": name, "url": url},
		),
	}
}

func init() {
	prometheus.MustRegister(version.NewCollector("alauda_monitor_exporter"))
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

	//begin
	phoenix := NewServiceDiagnoser("phoenix", "https://phoenix.alauda.cn/_diagnose")
	reg := prometheus.NewPedanticRegistry()
	reg.MustRegister(phoenix)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Alauda Monitor Exporter</title></head>
			<body>
			<h1>Alauda Monitor Exporter</h1>
			<p><a href="` + *metricsPath + `">Metrics</a></p>
			</body>
			</html>`))
	})

	gatherers := prometheus.Gatherers{
		prometheus.DefaultGatherer,
		reg,
	}

	h := promhttp.HandlerFor(gatherers,
		promhttp.HandlerOpts{
			ErrorLog:      log.NewErrorLogger(),
			ErrorHandling: promhttp.ContinueOnError,
		})
	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	})
	log.Infoln("Listening on", *listenAddress)
	if err := http.ListenAndServe(*listenAddress, nil); err != nil {
		log.Errorf("Error occur when start server %v", err)
		os.Exit(1)
	}

}
