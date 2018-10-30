package test

import (
	"fmt"
	"time"
	"math/rand"
)

func main() {
	/* 定义局部变量 */
	var a int = 100
	var b int= 200

	fmt.Printf("交换前 a 的值 : %d\n", a )
	fmt.Printf("交换前 b 的值 : %d\n", b )

	/* 调用函数用于交换值
	* &a 指向 a 变量的地址
	* &b 指向 b 变量的地址
	*/
	swap(&a, &b);

	fmt.Printf("交换后 a 的值 : %d\n", a )
	fmt.Printf("交换后 b 的值 : %d\n", b )

	for i := 0; i < 1; i++ {
		go func() {
			for {
				fmt.Println(i)
				time.Sleep((time.Duration)(1000) * time.Millisecond)
			}
		}()
	}
	fmt.Println(rand.Float64())
}
//a:xaaaa   100
//b:xcccc   200


func swap(x *int, y *int) {
	var temp int
	temp = *x    /* 保存 x 地址的值 */
	fmt.Print(x, "\n")
	fmt.Println(temp, "\n")
	//fmt.Println(*y)
	*x = *y      /* 将 y 赋值给 x */
	*y = temp    /* 将 temp 赋值给 y */
}

func swap2(x *int, y *int) {
	fmt.Print(x, "\n")
	fmt.Print(y, "\n")
	var temp *int
	temp = x    /* 保存 x 地址的值 */
	x = y      /* 将 y 赋值给 x */
	y = temp    /* 将 temp 赋值给 y */
	fmt.Print(x, "\n")
	fmt.Print(y, "\n")
}