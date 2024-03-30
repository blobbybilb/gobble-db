package main

import (
	"encoding/gob"
	"os"
)

type Person struct {
	Name string
	Age  int
	Str1 string
	Str2 string
	Str3 string
	Str4 string
	Str5 string
	Str6 string
}

func main() {
	p := Person{"Alice", 30, "some really \nlong string", "some really long string", "jksadfhkaljsfhkashfajkdfaklfhdlkjasfdh", "some really long string", "some really long string", "some really long string"}

	// Create an encoder and write to test2.gob
	file, _ := os.Create("test2.gob")
	enc := gob.NewEncoder(file)
	enc.Encode(p)
	file.Close()

	// Create a decoder and read from test2.gob
	file, _ = os.Open("test2.gob")
	dec := gob.NewDecoder(file)
	var p2 Person
	dec.Decode(&p2)
	file.Close()

	// Print the decoded struct
	println(p2.Name, p2.Age)

}
