package test

import (
	"fmt"
	//"reflect"
	"net/http"
	"io/ioutil"
	//"github.com/bitly/go-simplejson"
	"encoding/json"
	"time"
)

var componenetList1 = map[string] string{
	"furion": "https://phoenix.alauda.cn/_diagnose",
	"jakiro": "http://k8s-jakiro",
}

func HttpGet1(url string) (*map[string]interface{},error){
	defer func() {
		fmt.Println("b")
		if err := recover();err != nil {
			fmt.Println(err)
			dm1 := make(map[string]interface{})
			dm1["status"] = 1
		}
		fmt.Println("d")
	}()
	dm := make(map[string]interface{})
	httpCLi := &http.Client{
		Timeout: 15 * time.Second,
	}
	res, err := httpCLi.Get(url)
	defer res.Body.Close()
	bytessa, _:= ioutil.ReadAll(res.Body)
	fmt.Println(string(bytessa))
	json.Unmarshal(bytessa, &dm)
	return &dm, err
}


func Diagnose() {
	defer func() {
		fmt.Println("b")
		if err := recover();err != nil {
			fmt.Println(err)
			dm1 := make(map[string]interface{})
			dm1["status"] = 1
		}
		fmt.Println("d")
	}()
	for comp, url := range componenetList1{

		fmt.Sprintf("begin request %s", comp)
		res, err := HttpGet1(url)
		if err != nil {
			fmt.Println(err)
		}
		hahah := *res//make(map[string]interface{})
		//hahah =
		sssa, _ := json.Marshal(hahah["details"])
		fmt.Println(string(sssa))
	}

}


func main() {
	Diagnose()
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
