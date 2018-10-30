package test
import "fmt"
import "os"
import "time"
import "strconv"

/*
go run helloworld.go [max] [wait(0|1)]
e.g.
go run helloworld.go 5 0
go run helloworld.go 100 0
go run helloworld.go 100 1
*/

func foo(id, max int) {
	for i := 1; i <= max; i++ {
		fmt.Println(id, i)
	}
}

func waitKey() {
	var input string
	fmt.Scanln(&input)
	fmt.Println("Entered:", input)
}

func getParams() (max int, wait bool, ok bool) {
	if len(os.Args) != 3 {
		ok = false
		return
	}

	max, err := strconv.Atoi(os.Args[1])
	if err != nil {
		ok = false
		return
	}

	wait, err = strconv.ParseBool(os.Args[2])
	if err != nil {
		ok = false
		return
	}

	ok = true
	return
}


func main() {
	max, wait, ok := getParams()
	if !ok {
		fmt.Println("Usage: go run helloworld.go [max] [wait(0|1)]")
		return
	}

	for i := 1; i <= 3; i++ {
		go foo(i, max)
	}
	time.Sleep(3 * time.Second)
	if wait {
		waitKey()
	}
	time.Sleep(3 * time.Second)
}