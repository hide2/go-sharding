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
	for t := 0; t < GoShardingTableNumber; t++ {
		sql := fmt.Sprintf("DROP TABLE IF EXISTS event_%d", t)
		fmt.Println("[Exec]", sql)
		User.Exec(sql)
	}
	// Create Sharding tables
	fmt.Println("[CreateTable]")
	User.CreateTable()
	Event.CreateTable()

	// C
	var uid, uid2, uid3 int64
	for i := 0; i < 10; i++ {
		u := User.New()
		u.Name = "John"
		u.CreatedAt = time.Now()
		u.Save()
		fmt.Println("[Save User]", u.ID, u.Uid, u.Name, u.Datasource, u.Table)

		if i == 0 {
			uid = u.Uid
		}
		if i == 1 {
			uid2 = u.Uid
		}
		if i == 2 {
			uid3 = u.Uid
		}
	}
	fmt.Println(uid, uid2, uid3)
	for i := 0; i < 10; i++ {
		e := Event.New()
		e.Uid = uid3
		e.Event = "buy"
		e.CreatedAt = time.Now()
		e.Save()
		fmt.Println("[Save Event]", e.ID, e.Uid, e.Event, e.Datasource, e.Table)
	}

	// R
	u, e := User.FindByUid(123)
	fmt.Println("[FindByUid(123)]", u, e)
	u, _ = User.FindByUid(uid)
	fmt.Println("[FindByUid]", u.ID, u.Uid, u.Name, u.Datasource, u.Table)
	u, _ = User.FindByUidAndID(uid, 1)
	fmt.Println("[FindByUidAndID]", u.ID, u.Uid, u.Name, u.Datasource, u.Table)

	// U
	u.Name = "Calvin"
	u.Save()
	fmt.Println("[Update]", u.ID, u.Uid, u.Name, u.Datasource, u.Table)

	u, _ = User.FindByUid(uid)
	fmt.Println("[FindByUid]", u.ID, u.Uid, u.Name, u.Datasource, u.Table)

	// D
	u, _ = User.FindByUid(uid)
	fmt.Println("[Before Delete FindByUid]", u)
	u.Delete()
	u, _ = User.FindByUid(uid)
	fmt.Println("[After Delete FindByUid]", u)

	u, _ = User.FindByUid(uid2)
	fmt.Println("[Before Destroy FindByUid]", u)
	User.DestroyByUid(uid2)
	u, _ = User.FindByUid(uid2)
	fmt.Println("[After Destroy FindByUid]", u)

	// Create
	for i := 0; i < 10; i++ {
		props := map[string]interface{}{"name": "Dog", "created_at": time.Now()}
		u, _ = User.Create(props)
		fmt.Println("[Create User]", u.ID, u.Uid, u.Name, u.Datasource, u.Table)
	}
	for i := 0; i < 10; i++ {
		props := map[string]interface{}{"uid": uid3, "event": "gold", "created_at": time.Now()}
		e, _ := Event.Create(props)
		fmt.Println("[Create Event]", e.ID, e.Uid, e.Event, e.Datasource, e.Table)
	}

	// WHERE
	conds := map[string]interface{}{"uid": uid3}
	es, _ := Event.Where(conds)
	for _, v := range es {
		fmt.Println("[Where Sid]", v.ID, v.Uid, v.Event, v.Datasource, v.Table)
	}
	// WHERE OrderBy&Limit&Offset
	conds = map[string]interface{}{"uid": uid3}
	es, _ = Event.OrderBy("ID DESC").Offset(2).Limit(3).Where(conds)
	for _, v := range es {
		fmt.Println("[Where Sid Limit Offset OrderBy]", v.ID, v.Uid, v.Event, v.Datasource, v.Table)
	}
	// WHERE Paginate
	conds = map[string]interface{}{"uid": uid3}
	es, _ = Event.Page(1, 5).Where(conds)
	for _, v := range es {
		fmt.Println("[Where Paginate]", v.ID, v.Uid, v.Event, v.Datasource, v.Table)
	}
	// WHERE with no Sharding Column, Find Across All Databases & Tables
	conds = map[string]interface{}{"event": "gold"}
	es, _ = Event.Where(conds)
	for _, v := range es {
		fmt.Println("[Where No Sid]", v.ID, v.Uid, v.Event, v.Datasource, v.Table)
	}

	// COUNT
	conds2 := map[string]interface{}{"uid": uid3}
	conds3 := map[string]interface{}{"uid": uid3, "event": "gold"}
	conds4 := map[string]interface{}{"event": "gold"}
	c, _ := Event.CountAll()
	fmt.Println("[CountAll]", c)
	c, _ = Event.Count(conds2)
	fmt.Println("[Count With Sid]", c)
	c, _ = Event.Count(conds3)
	fmt.Println("[Count With Sid And Conds]", c)
	c, _ = Event.Count(conds4)
	fmt.Println("[Count Without Sid]", c)

	// UPDATE
	props3 := map[string]interface{}{"event": "item"}
	conds5 := map[string]interface{}{"uid": uid3, "event": "item"}
	Event.Update(props3, conds3)
	es2, _ := Event.Where(conds5)
	fmt.Println("[Update]", len(es2))
	for _, v := range es2 {
		fmt.Println("[Update]", v.ID, v.Uid, v.Event, v.Datasource, v.Table)
	}
}
