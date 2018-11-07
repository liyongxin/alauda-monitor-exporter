
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
	"sync"
)


type GlobalServiceDiagnose struct{
	ServiceName string
	DiagnoseUrl string
	HealthDesc  *prometheus.Desc
}

type diagnoseRes struct {
	Name string
	Url string
	Code float64
	Status string  `json:"status"`
	Details []map[string]string `json:"details"`
	DetailStr string
}

var  CompList  = map[string]string {
	"furion" : "https://furion.alauda.cn:8443/_diagnose",
	"phoenix": "https://phoenix.alauda.cn/_diagnose",
}

// Describe simply sends Descs in the struct to the channel.
func (g *GlobalServiceDiagnose) Describe(ch chan<- *prometheus.Desc) {
	ch <- g.HealthDesc
}

func (g *GlobalServiceDiagnose) Collect(ch chan<- prometheus.Metric) {
	diagRes := g.healthCheck()
	for _, res := range *diagRes {
		ch <- prometheus.MustNewConstMetric(
			g.HealthDesc,
			prometheus.GaugeValue,
			res.Code,
			res.Status,
			res.DetailStr,
		)
	}
}

/*
{"status":"OK","details":[{"status":"OK","name":"DATABASE"}]}
*/
func (g *GlobalServiceDiagnose) healthCheck ()  *[]diagnoseRes{
	var wg sync.WaitGroup
	resCh := make(chan diagnoseRes, len(CompList))

	for compName, compUrl := range CompList {
		go HttpGet(compName, compUrl, &wg, &resCh)
		wg.Add(1)
	}

	wg.Wait()
	close(resCh)

	var res  []diagnoseRes

	for diagRes := range resCh {
		if diagRes.Status == "DANGER" {
			diagRes.Code = -1
		}else if diagRes.Status == "ERROR"{
			diagRes.Code = 0
		}else {
			diagRes.Code = 1
		}
		res = append(res, diagRes)
	}
	return &res
}

func HttpGet(name, url string, wg *sync.WaitGroup, ch *chan diagnoseRes) {
	res := diagnoseRes{}
	defer wg.Done()
	defer func() {
		if err := recover();err != nil {
			errMsg := fmt.Sprintf("get health check error,%v", err)
			log.Errorln(errMsg)
			res.Name = name
			res.Status = "DANGER"
			res.DetailStr = errMsg
			*ch <- res
		}
	}()

	httpCLi := &http.Client{
		Timeout: 4 * time.Second,
	}
	resp, err := httpCLi.Get(url)
	if resp == nil && err != nil{
		errMsg := fmt.Sprintf("get health check error,%v", err)
		log.Errorln(errMsg)
		res.Name = name
		res.Status = "DANGER"
		res.DetailStr = errMsg
		*ch <- res
		return
	}
	defer resp.Body.Close()
	if err != nil {
		log.Errorln(err)
	}
	bts, _:= ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(bts, &res)
	if err != nil {
		log.Errorln(err)
	}
	res.Name = name
	res.DetailStr = string(bts)
	*ch <- res
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
