package test

import "fmt"


func Init(num1 string, num2 int) (string, int) {
	/* Heek */
	var a, b = "hello", "world"
	fmt.Println(&a)
	fmt.Println(b, a)
	return num1, num2
}

func main() {
	a := "aa"
	b := 2
	c, d := Init(a, b)
	fmt.Printf("value a is %s, b is %d\n", c, d)
	fmt.Println("Hello, World!")
}
