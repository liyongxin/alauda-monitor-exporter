package test

import "fmt"

type Animal interface {
	print() string
}

type Dog struct {
	name string
	age  int
}

type Snake struct {
	name string
	age int
	action string
}

func (dog Dog) print() string  {
	fmt.Println(dog.name)
	return dog.name
}

func (snake Snake) print() string {
	fmt.Println(snake.name)
	return snake.name
}

func main() {
	var ani  Animal
	ani = Snake{
		name: "snake1",
	}
	ani.print()

	ani = Dog{
		name: "dog1",
	}
	ani.print()

	if myName := ani.print(); myName != "" {
		fmt.Printf("myName is %s", myName)
	}else {
		fmt.Println("myName is null")
	}
}