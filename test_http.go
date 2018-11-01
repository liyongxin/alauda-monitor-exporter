package main

import (
	"fmt"
	//"reflect"
	"time"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"github.com/prometheus/common/log"
)

var componenetList1 = map[string] string{
	"jakiro": "https://phoenix.alauda.cn/_diagnose",
	"furion": "https://furion1.alauda.cn1:8443/_diagnose",
}


type diagnoseRes1 struct {
	status string
	details []map[string]string
	detailStr string
}

type diagnoseMsg1 struct {
	status float64
	details string
}

func HttpGet1(url string)  (res *diagnoseRes1){
	res = &diagnoseRes1{}
	defer func() {
		if err := recover();err != nil {
			errMsg := fmt.Sprintf("get health check error,%v", err)
			//log.Errorln(errMsg)
			res = &diagnoseRes1{
				status: "DANGER",
				detailStr: errMsg,
			}
		}
	}()

	httpCLi := &http.Client{
		Timeout: 5 * time.Second,
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
	log.Errorln(res)
	return res
}

func Diagnose() {

	for comp, url := range componenetList1{

		fmt.Println(fmt.Sprintf("begin request %s", comp))
		fmt.Println(fmt.Sprintf("begin request url %s", url))
		res := HttpGet1(url)


		hahah := *res//make(map[string]interface{})
		//hahah =
		//sssa, _ := json.Marshal(hahah["details"])
		//fmt.Println(hahah)
		log.Errorln(hahah)
	}

}

type yxli struct{
	name string
	age int
}

func test()  {
	res := getReturn("liyx", 11)
	fmt.Println(*res)
}

func getReturn(name string, age int) (yx *yxli) {
	yx = &yxli{}
	yx.name = name
	yx.age = age
	return yx
}

func main() {
	Diagnose()
	//test()
	/*data, err := simplejson.NewJson(body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(*data)
	//sss, err:= data.Get("details").Array()
	status, err:= data.Get("details").String()
	//fmt.Println(len(sss))
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(status)
	//fmt.Println(reflect.TypeOf(*data.Get("status")).String())
	*/
}
