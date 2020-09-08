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

# Install Go
```
sudo rm -fr /usr/local/go
Download & Install MacOS pkg from https://golang.org/dl/
export PATH=$PATH:/usr/local/go/bin
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

sql_log: false
slow_sql_log: 2

sharding_table_number: 256
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
```
go run gen.go
```
Which will generate Model files
