package main

import (
	"fmt"
	"time"

	. "github.com/hide2/go-sharding/db"
	. "github.com/hide2/go-sharding/model"
)

func main() {
	// Gen UUID
	for i := 0; i < 10; i++ {
		uid := GenUUID()
		fmt.Println("= Gen", uid)
	}
	// Drop Sharding tables
	for t := 0; t < GoShardingTableNumber; t++ {
		User.Exec(fmt.Sprintf("DROP TABLE IF EXISTS user_%d", t))
	}
	// Create Sharding tables
	User.CreateTable()

	// C
	for i := 0; i < 100; i++ {
		u := User.New()
		u.Name = "John"
		u.CreatedAt = time.Now()
		u.Save()
		ds_fix := int64(u.Uid) / int64(GoShardingTableNumber) % int64(GoShardingDatasourceNumber)
		table_fix := int64(u.Uid) % int64(GoShardingTableNumber)
		fmt.Println("[Save]", u.ID, u.Uid, ds_fix, table_fix)
	}
}
