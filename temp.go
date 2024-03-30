package main

// import (
// 	"fmt"
// 	"reflect"
// )

// type Person2 struct {
// 	Name string
// 	Age  int
// }

// func main2() {
// 	p := Person{"Alice", 30}

// 	// Inspect the type and value
// 	t := reflect.TypeOf(p)
// 	v := reflect.ValueOf(p)

// 	fmt.Println("Type:", t)
// 	fmt.Println("Value:", v)

// 	x := map[interface{}]int{"Alice": 30, []interface{}{}: 25}
// 	fmt.Println(x)
// 	// Iterate over struct fields
// 	for i := 0; i < t.NumField(); i++ {
// 		field := t.Field(i)
// 		value := v.Field(i)
// 		fmt.Println(field.Name, ":", value)
// 	}
// }
