package main

import (
	"fmt"
)

// DBTable struct
// Methods: Insert, Update, Delete, Select

type DBTable[T any] struct {
	Name string
}

type Query[T any] func(T) bool

func (t *DBTable[T]) Insert(data T) {
	fmt.Println("Inserting", data, "into", t.Name)
}

func (t *DBTable[T]) Update(query Query[T], data T) {
	fmt.Println("Updating", data, "in", t.Name)
}

func (t *DBTable[T]) Delete(query Query[T]) {
	fmt.Println("Deleting", query, "from", t.Name)
}

func (t *DBTable[T]) Select(query Query[T]) []T {
	fmt.Println("Selecting", query, "from", t.Name)
	return []T{}
}

func (t *DBTable[T]) GetID() int {
	return 0
}
