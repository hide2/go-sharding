Go-Sharding is a simple Sharding middleware based on GO-ORM(https://github.com/hide2/go-orm)

# Go-Sharding Features
- Auto mapping SQL to Sharding databases & tables
- Distributed UUID generator based on Snowflake
- Scaling databases & tables
- Auto create table scripts
- Model & CRUD methods generator
- Connection Pool
- Write/Read Splitting
- SQL log & Slow SQL log for profiling

# Install Go on Mac
``` bash
sudo rm -fr /usr/local/go
Download & Install MacOS pkg from https://golang.org/dl/
export PATH=$PATH:/usr/local/go/bin
```

# Install Go on CentOS
``` bash
wget https://golang.org/dl/go1.15.1.linux-amd64.tar.gz
sudo tar -C /usr/local/ -xzvf go1.15.1.linux-amd64.tar.gz
sudo vi /etc/profile
export GOROOT=/usr/local/go
export GOPATH=/data/go
export GOBIN=$GOPATH/bin
export PATH=$PATH:$GOROOT/bin
export PATH=$PATH:$GOPATH/bin
```

# Usage
Define Datasources in datasource.yml
``` yml
datasources:
  - name: ds_0
    write: root:root@tcp(127.0.0.1:3306)/my_db_0?charset=utf8mb4&parseTime=True
    read: root:root@tcp(127.0.0.1:3306)/my_db_0?charset=utf8mb4&parseTime=True

  - name: ds_1
    write: root:root@tcp(127.0.0.1:3306)/my_db_1?charset=utf8mb4&parseTime=True
    read: root:root@tcp(127.0.0.1:3306)/my_db_1?charset=utf8mb4&parseTime=True
  
  - name: ds_2
    write: root:root@tcp(127.0.0.1:3306)/my_db_2?charset=utf8mb4&parseTime=True
    read: root:root@tcp(127.0.0.1:3306)/my_db_2?charset=utf8mb4&parseTime=True

  - name: ds_3
    write: root:root@tcp(127.0.0.1:3306)/my_db_3?charset=utf8mb4&parseTime=True
    read: root:root@tcp(127.0.0.1:3306)/my_db_3?charset=utf8mb4&parseTime=True

  - name: ds_4
    write: root:root@tcp(127.0.0.1:3306)/my_db_4?charset=utf8mb4&parseTime=True
    read: root:root@tcp(127.0.0.1:3306)/my_db_4?charset=utf8mb4&parseTime=True

  - name: ds_5
    write: root:root@tcp(127.0.0.1:3306)/my_db_5?charset=utf8mb4&parseTime=True
    read: root:root@tcp(127.0.0.1:3306)/my_db_5?charset=utf8mb4&parseTime=True

    sql_log: false
slow_sql_log: 500

sharding_table_number: 255
sharding_column: uid
sharding_node_id: 1 (0~1023)
```
Define Models in model.yml
``` yml
models:
  - model: User
    uid: int64|auto
    name: string
    created_at: time.Time

  - model: Event
    uid: int64
    event: string
    created_at: time.Time
```
Generate Model go files
``` bash
go run gen.go
```
You can use your Model for Sharding databases & tables now:
``` go
package main

import (
	"fmt"

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
	fmt.Println("[CreateTable]")
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
	u, e := User.FindByUid(123)
	fmt.Println("[FindByUid(123)]", u, e)
	u, _ = User.FindByUid(uid)
	fmt.Println("[FindByUid]", u.ID, u.Uid, u.Name, u.Datasource, u.Table)

	// U
	u.Name = "Calvin"
	u.Save()
	fmt.Println("[Update]", u.ID, u.Uid, u.Name, u.Datasource, u.Table)

	u, _ = User.FindByUid(uid)
	fmt.Println("[FindByUid]", u.ID, u.Uid, u.Name, u.Datasource, u.Table)
}

```