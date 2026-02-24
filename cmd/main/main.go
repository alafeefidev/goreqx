package main

import (
	"fmt"

	"github.com/google/uuid"
)

var sus uuid.UUIDs

func main() {
	fmt.Println(uuid.New())
}