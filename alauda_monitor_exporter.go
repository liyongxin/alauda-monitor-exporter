
package main

import (
	"net/http"
	_ "net/http/pprof"
	//"math/rand"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"
	"io/ioutil"
	"encoding/json"
	"fmt"
	"os"
	//"strconv"
	"time"
)


type GlobalServiceDiagnose struct{
	serviceName string
	diagnoseUrl string
	healthDesc  *prometheus.Desc
}

type diagnoseKVPair map[string]interface{}
/*var componenetList = map[string] string{
	"furion": "http://k8s-furion:8080",
	"jakiro": "http://k8s-jakiro",
}*/

// Describe simply sends the two Descs in the struct to the channel.
func (g *GlobalServiceDiagnose) Describe(ch chan<- *prometheus.Desc) {
	ch <- g.healthDesc
}

func (g *GlobalServiceDiagnose) Collect(ch chan<- prometheus.Metric) {
	diagMsg := g.healthCheck()
//	value,_:=diag["health"].(string)
	//	health , err := strconv.ParseFloat(value, 64)

	ch <- prometheus.MustNewConstMetric(
		g.healthDesc,
		prometheus.GaugeValue,
		diagMsg.status,
		g.serviceName,
		diagMsg.details,
	)
}

/*
{
	"furion": {"health": 1, "message": "everything is ok"},
	"jakiro": {"health": 0, "message": "everything is bad"}
}
*/
func (g *GlobalServiceDiagnose) healthCheck ()  *diagnoseMsg{
	diagRes := HttpGet(g.diagnoseUrl)
	res := &diagnoseMsg{}
	res.status = 0
	if diagRes.status == "OK" {
		res.status = 1
	}
	res.details = diagRes.detailStr
	return res
}

type diagnoseRes struct {
	status string
	details []map[string]string
	detailStr string
}

type diagnoseMsg struct {
	status float64
	details string
}

func HttpGet(url string)  (res *diagnoseRes){
	res = &diagnoseRes{}
	defer func() {
		if err := recover();err != nil {
			errMsg := fmt.Sprintf("get health check error,%v", err)
			log.Errorln(errMsg)
			res = &diagnoseRes{
				status: "DANGER",
				detailStr: errMsg,
			}
		}
	}()

	httpCLi := &http.Client{
		Timeout: 15 * time.Second,
	}
	resp, err := httpCLi.Get(url)
	defer resp.Body.Close()
	if err != nil {
		log.Errorln(err)
	}
	bts, _:= ioutil.ReadAll(resp.Body)

	log.Debugln(string(bts))
	err = json.Unmarshal(bts, res)
	if err != nil {
		log.Errorln(err)
	}
	res.detailStr = string(bts)
	return res
}

func NewServiceDiagnoser(name, url string) *GlobalServiceDiagnose{
	return &GlobalServiceDiagnose{
		serviceName: name,
		diagnoseUrl: url,
		healthDesc: prometheus.NewDesc(
			fmt.Sprintf("global_service_diagnose_%s",name),
			fmt.Sprintf("global service %s diagnose.",name),
			[]string{"path", "message"},
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
	jakiro := NewServiceDiagnoser("jakiro", "https://api.alauda.cn/_diagnose")
	//furion := NewServiceDiagnoser("furion", "https://furion.alauda.cn:8443/_diagnose")
	reg := prometheus.NewPedanticRegistry()
	reg.MustRegister(phoenix)
	reg.MustRegister(jakiro)
	//reg.MustRegister(furion)
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
