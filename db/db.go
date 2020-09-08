package db

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/hide2/go-sharding/snowflake"
	"gopkg.in/yaml.v2"
)

type Datasources struct {
	Datasources        []Datasource `yaml:"datasources,flow"`
	SqlLog             bool         `yaml:"sql_log"`
	SlowSqlLog         int          `yaml:"slow_sql_log"`
	ShardingTableNumer int          `yaml:"sharding_table_number"`
	ShardingColumn     string       `yaml:"sharding_column"`
	ShardingNodeId     int64        `yaml:"sharding_node_id"`
}

type Datasource struct {
	Name  string `yaml:"name"`
	Write string `yaml:"write"`
	Read  string `yaml:"read"`
}

var DBPool = make(map[string]map[string]*sql.DB)

var GoOrmSqlLog = false
var GoOrmSlowSqlLog = 0
var GoShardingTableNumer int
var GoShardingColumn string
var GoShardingNodeId int64
var node *snowflake.Node

// Init DBPool
func init() {
	f, _ := ioutil.ReadFile("datasource.yml")
	dss := Datasources{}
	err := yaml.Unmarshal(f, &dss)
	if err != nil {
		fmt.Println("error:", err)
	}
	GoOrmSqlLog = dss.SqlLog
	GoOrmSlowSqlLog = dss.SlowSqlLog
	GoShardingTableNumer = dss.ShardingTableNumer
	GoShardingColumn = dss.ShardingColumn
	GoShardingNodeId = dss.ShardingNodeId
	for _, ds := range dss.Datasources {
		wdb, err := sql.Open("mysql", ds.Write)
		if err != nil {
			fmt.Println("Connection to mysql failed:", err)
			return
		}
		wdb.SetConnMaxLifetime(100 * time.Second) //最大连接周期，超时的连接就close
		wdb.SetMaxOpenConns(100)                  //设置最大连接数
		rdb, err := sql.Open("mysql", ds.Read)
		if err != nil {
			fmt.Println("Connection to mysql failed:", err)
			return
		}
		rdb.SetConnMaxLifetime(100 * time.Second) //最大连接周期，超时的连接就close
		rdb.SetMaxOpenConns(100)                  //设置最大连接数

		DBPool[ds.Name] = make(map[string]*sql.DB)
		DBPool[ds.Name]["w"] = wdb
		DBPool[ds.Name]["r"] = rdb
	}

	fmt.Println("=== Datasource initialized!")
	for i, ds := range dss.Datasources {
		fmt.Println("DS", i, ds.Name)
		fmt.Println("Write", ds.Write)
		fmt.Println("Read", ds.Read)
	}
	fmt.Println("SqlLog", dss.SqlLog)
	fmt.Println("SlowSqlLog", dss.SlowSqlLog)
	fmt.Println("ShardingTableNumer", dss.ShardingTableNumer)
	fmt.Println("ShardingColumn", dss.ShardingColumn)
	fmt.Println("ShardingNodeId", dss.ShardingNodeId)

	node, err = snowflake.NewNode(GoShardingNodeId)
	if err != nil {
		fmt.Printf("Error creating NewNode, %s\n", err)
	} else {
		fmt.Printf("=== Snowflake Node %d initialized!\n", GoShardingNodeId)
	}
}

func GenUUID() int64 {
	return node.Generate().Int64()
}
