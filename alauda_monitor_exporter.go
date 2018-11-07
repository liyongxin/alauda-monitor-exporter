
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
	"time"
)


type GlobalServiceDiagnose struct{
	ServiceName string
	DiagnoseUrl string
	HealthDesc  *prometheus.Desc
}

type diagnoseRes struct {
	Code float64
	Status string  `json:"status"`
	Details []map[string]string `json:"details"`
	DetailStr string
}

var  CompList  = map[string]string {
	"furion" : "https://furion.alauda.cn:8443/_diagnose",
	"phoenix": "https://phoenix.alauda.cn/_diagnose",
	"architect": "http://architect.alauda.cn:8080/_diagnose",
	"windranger": "https://windranger.alauda.cn/_diagnose",
}

// Describe simply sends Descs in the struct to the channel.
func (g *GlobalServiceDiagnose) Describe(ch chan<- *prometheus.Desc) {
	ch <- g.HealthDesc
}

func (g *GlobalServiceDiagnose) Collect(ch chan<- prometheus.Metric) {
	diagMsg := g.healthCheck()

	ch <- prometheus.MustNewConstMetric(
		g.HealthDesc,
		prometheus.GaugeValue,
		diagMsg.Code,
		diagMsg.Status,
		diagMsg.DetailStr,
	)
}

/*
{"status":"OK","details":[{"status":"OK","name":"DATABASE"}]}
*/
func (g *GlobalServiceDiagnose) healthCheck ()  *diagnoseRes{
	diagRes := HttpGet(g.DiagnoseUrl)
	if diagRes.Status == "DANGER" {
		diagRes.Code = 0
	}else if  diagRes.Status == "EERROR" {
		diagRes.Code = 0
	}else {
		diagRes.Code = 1
	}
	return diagRes
}


func HttpGet(url string)  (res *diagnoseRes){
	res = &diagnoseRes{}
	defer func() {
		if err := recover();err != nil {
			errMsg := fmt.Sprintf("get health check error,%v", err)
			log.Errorln(errMsg)
			res.Status = "DANGER"
			res.DetailStr = errMsg
			return
		}
	}()

	httpCLi := &http.Client{
		Timeout: 4 * time.Second,
	}
	resp, err := httpCLi.Get(url)
	if resp == nil && err != nil{
		errMsg := fmt.Sprintf("get health check error,%v", err)
		log.Errorln(errMsg)
		res.Status = "DANGER"
		res.DetailStr = errMsg
		return res
	}
	defer resp.Body.Close()
	if err != nil {
		log.Errorln(err)
	}
	bts, _:= ioutil.ReadAll(resp.Body)

	err = json.Unmarshal(bts, res)
	if err != nil {
		log.Errorln(err)
	}
	res.DetailStr = string(bts)
	return res
}



func registryDiagnoser(name, url string) *GlobalServiceDiagnose{
	return &GlobalServiceDiagnose{
		ServiceName: name,
		DiagnoseUrl: url,
		HealthDesc: prometheus.NewDesc(
			"global_service_diagnose",
			"global service diagnose",
			[]string{"status", "details"},
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

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Alauda Monitor Exporter</title></head>
			<body>
			<h1>Alauda Monitor Exporter</h1>
			<p><a href="` + *metricsPath + `">Metrics</a></p>
			</body>
			</html>`))
	})

	reg := prometheus.NewPedanticRegistry()

	for name, url := range CompList{
		reg.MustRegister(registryDiagnoser(name, url))
	}

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
