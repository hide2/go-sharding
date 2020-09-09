package main

import (
	"fmt"

	. "github.com/hide2/go-sharding/db"
	. "github.com/hide2/go-sharding/model"
)

func main() {
	for i := 0; i < 20; i++ {
		fmt.Println("=", GenUUID())
	}

	User.CreateTable()
}
