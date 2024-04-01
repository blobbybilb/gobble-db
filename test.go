package main

import (
	"fmt"
	"os"
)

type Person struct {
	Name string
	Age  int
}

func verifyItemsEqual(ps ...[]Person) bool {
	if len(ps) == 0 {
		fmt.Println("No items")
		return true
	}

	for i := 1; i < len(ps); i++ {
		if len(ps[i-1]) != len(ps[i]) {
			fmt.Println("Lengths do not match")
			return false
		}

		for j := 0; j < len(ps[i-1]); j++ {
			if ps[i-1][j] != ps[i][j] {
				fmt.Println("FALSE")
				return false
			}
		}
	}

	fmt.Println("TRUE")
	return true
}

func sortPersonsByName(ps []Person) []Person {
	for i := 0; i < len(ps); i++ {
		for j := i + 1; j < len(ps); j++ {
			if ps[i].Name > ps[j].Name {
				ps[i], ps[j] = ps[j], ps[i]
			}
		}
	}

	return ps
}

func test() {
	db, _ := OpenDB("testdb")
	c, _ := OpenCollection[Person](db, "testcollection")
	i1, _ := OpenIndex[Person, string](&c, func(p Person) string {
		return p.Name
	})
	i2, _ := OpenIndex[Person, int](&c, func(p Person) int {
		return p.Age
	})

	_ = c.Insert(Person{Name: "Person 1", Age: 1})
	_ = c.Insert(Person{Name: "Person 2", Age: 2})
	_ = c.Insert(Person{Name: "Person 3", Age: 3})
	_ = c.Insert(Person{Name: "Person 4", Age: 4})
	_ = c.Insert(Person{Name: "Person 5", Age: 5})

	x, _ := i1.Get("Person 1")
	y, _ := i2.Get(1)
	a, _ := c.Select(func(p Person) bool { return p.Name == "Person 1" })
	b, _ := c.Select(func(p Person) bool { return p.Age == 1 })
	verifyItemsEqual(x, y, a, b, []Person{{Name: "Person 1", Age: 1}})

	x, _ = c.Select(func(p Person) bool { return p.Age < 3 })
	verifyItemsEqual(sortPersonsByName(x), []Person{{Name: "Person 1", Age: 1}, {Name: "Person 2", Age: 2}})

	_ = c.Modify(func(p Person) bool { return p.Age == 1 }, func(p Person) Person { p.Age = 10; return p })
	x, _ = c.Select(func(p Person) bool { return p.Age == 10 })
	y, _ = i2.Get(10)
	a, _ = c.Select(func(p Person) bool { return p.Name == "Person 2" })
	b, _ = i1.Get("Person 2")
	verifyItemsEqual(x, y, []Person{{Name: "Person 1", Age: 10}})
	verifyItemsEqual(a, b, []Person{{Name: "Person 2", Age: 2}})

	_ = c.Delete(func(p Person) bool { return p.Age == 10 })
	x, _ = c.Select(func(p Person) bool { return p.Age == 10 })
	y, _ = i2.Get(10)
	verifyItemsEqual(x, y, []Person{})

	//	delete all but one
	_ = c.Delete(func(p Person) bool { return p.Age > 2 })
	x, _ = c.Select(func(p Person) bool { return true })
	y, _ = i2.Get(2)
	verifyItemsEqual(x, y, []Person{{Name: "Person 2", Age: 2}})

	d, _ := c.Number()
	e, _ := i2.Num(2)
	fmt.Println(d == e)

	_ = c.Delete(func(p Person) bool { return true })
	x, _ = c.Select(func(p Person) bool { return true })
	//fmt.Println(i2.Index)
	//fmt.Println(i1.Index)
	//fmt.Println(x)

	// Re-insert the data
	_ = c.Insert(Person{Name: "Person 1", Age: 1})
	_ = c.Insert(Person{Name: "Person 2", Age: 2})
	_ = c.Insert(Person{Name: "Person 3", Age: 3})
	_ = c.Insert(Person{Name: "Person 4", Age: 4})
	_ = c.Insert(Person{Name: "Person 5", Age: 5})

	// Test the index methods
	_ = i1.Del("Person 1")
	_ = i2.Del(2)
	_ = i1.Mod("Person 3", func(p Person) Person { p.Name = "Person 3 2"; return p })
	_ = i2.Mod(4, func(p Person) Person { p.Age = 40; return p })

	x, _ = i1.Get("Person 1")
	y, _ = i2.Get(2)
	a, _ = i1.Get("Person 3 2")
	b, _ = i2.Get(40)
	verifyItemsEqual(x, []Person{}, y, []Person{})
	verifyItemsEqual(a, []Person{{Name: "Person 3 2", Age: 3}})
	verifyItemsEqual(b, []Person{{Name: "Person 4", Age: 40}})

	x, _ = c.Select(func(p Person) bool { return true })
	fmt.Println(x)
}

func main() {
	test()
	bench()
	_ = os.RemoveAll("testdb")
	_ = os.RemoveAll("benchdb")
}
