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
		fmt.Println("[GenUUID]", uid)
	}
	// Drop Sharding tables
	for t := 0; t < GoShardingTableNumber; t++ {
		sql := fmt.Sprintf("DROP TABLE IF EXISTS user_%d", t)
		fmt.Println("[Exec]", sql)
		User.Exec(sql)
	}
	// Create Sharding tables
	fmt.Sprintf("[CreateTable]")
	User.CreateTable()

	// C
	var uid int64
	for i := 0; i < 10; i++ {
		u := User.New()
		u.Name = "John"
		u.CreatedAt = time.Now()
		u.Save()
		fmt.Println("[Save]", u.ID, u.Uid, u.Datasource, u.Table)
		uid = u.Uid
	}

	// R
	u, e := User.FindByUid(uid)
	fmt.Println("[FindByUid] DS=ds_%d Table=user_%d", int64(u.Uid)/int64(GoShardingTableNumber)%int64(GoShardingDatasourceNumber), e)
	fmt.Println("[FindByUid]", u, e)
}
