
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


// Describe simply sends the two Descs in the struct to the channel.
func (g *GlobalServiceDiagnose) Describe(ch chan<- *prometheus.Desc) {
	ch <- g.HealthDesc
}

func (g *GlobalServiceDiagnose) Collect(ch chan<- prometheus.Metric) {
	diagMsg := g.healthCheck()
//	value,_:=diag["health"].(string)
	//	health , err := strconv.ParseFloat(value, 64)

	ch <- prometheus.MustNewConstMetric(
		g.HealthDesc,
		prometheus.GaugeValue,
		diagMsg.Code,
		g.ServiceName,
		diagMsg.Status,
	)
}

/*
{
	"furion": {"health": 1, "message": "everything is ok"},
	"jakiro": {"health": 0, "message": "everything is bad"}
}
*/
func (g *GlobalServiceDiagnose) healthCheck ()  *diagnoseMsg{
	diagRes := HttpGet(g.DiagnoseUrl)
	log.Errorln(diagRes)
	res := &diagnoseMsg{}
	res.Code = 0
	if diagRes.Status == "OK" {
		res.Code = 1
	}
	res.Status = diagRes.Status
	res.Details = diagRes.DetailStr
	return res
}

type diagnoseRes struct {
	Status string
	Details []map[string]string
	DetailStr string
}

type diagnoseMsg struct {
	Code float64
	Status string
	Details string
}

func HttpGet(url string)  (res *diagnoseRes){
	res = &diagnoseRes{}
	defer func() {
		if err := recover();err != nil {
			errMsg := fmt.Sprintf("get health check error,%v", err)
			log.Errorln(errMsg)
			res.Status = "DANGER"
			res.DetailStr = errMsg
		}
	}()

	httpCLi := &http.Client{
		Timeout: 3 * time.Second,
	}
	resp, err := httpCLi.Get(url)
	if resp == nil && err != nil{
		log.Errorln(fmt.Sprintf("get health check error,%v", err))
		res.Status = "DANGER"
		res.DetailStr = fmt.Sprintf("get health check error,%v", err)
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

func NewServiceDiagnoser(name, url string) *GlobalServiceDiagnose{
	return &GlobalServiceDiagnose{
		ServiceName: name,
		DiagnoseUrl: url,
		HealthDesc: prometheus.NewDesc(
			fmt.Sprintf("global_service_diagnose_%s",name),
			fmt.Sprintf("global service %s diagnose.",name),
			[]string{"global_service_name", "status"},
			prometheus.Labels{"url": url},
		),
	}
}

func MetricCreator() *prometheus.Registry {
	phoenix := NewServiceDiagnoser("phoenix", "https://phoenix.alauda.cn/_diagnose")
	furion := NewServiceDiagnoser("furion", "https://furion.alauda.cn:8443/_diagnose")
	furion2 := NewServiceDiagnoser("furion2", "https://furion2.alauda.cn:8443/_diagnose")
	reg := prometheus.NewPedanticRegistry()
	reg.MustRegister(phoenix)
	reg.MustRegister(furion)
	reg.MustRegister(furion2)
	return reg
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
