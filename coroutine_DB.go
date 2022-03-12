package main

import (
	"fmt"
)

type Data struct {
}
type DataTable struct {
}
type Lock struct {
	mode int
}
type LockTable struct {
	Table []Lock
}

func main() {
	fmt.Println("My favorite number is 1232")
}
