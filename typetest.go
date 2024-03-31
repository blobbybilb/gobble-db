package main

type A[T any] struct {
	Val T
}

type B[T any, D any] struct {
	Val  A[T]
	Val2 D
}
