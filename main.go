package main

import (
	"fmt"

	. "github.com/hide2/go-sharding/db"
	. "github.com/hide2/go-sharding/model"
)

func main() {
	// Gen UUID
	for i := 0; i < 20; i++ {
		fmt.Println("=", GenUUID())
	}
	// Drop Sharding tables
	for t := 0; t < GoShardingTableNumber; t++ {
		User.Exec(fmt.Sprintf("DROP TABLE IF EXISTS user_%d", t))
	}
	// Create Sharding tables
	User.CreateTable()
}
