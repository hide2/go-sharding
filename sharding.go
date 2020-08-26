package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"gopkg.in/yaml.v2"
)

type ShardingConfig struct {
	Databases      []string `yaml:"datasources,flow"`
	TableNumber    string   `yaml:"table_number"`
	ShardingColumn string   `yaml:"sharding_column"`
}

func main() {
	f, _ := ioutil.ReadFile("sharding.yml")
	s := ShardingConfig{}
	err := yaml.Unmarshal(f, &s)
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Println(s)
	fmt.Println(len(s.Databases))
	fmt.Println(s.Databases[0])
	fmt.Println(s.Databases[1])
	fmt.Println(s.Databases[2])
	fmt.Println(s.Databases[3])
	fmt.Println(s.TableNumber)
	fmt.Println(s.ShardingColumn)

	conn := s.Databases[0]
	DB, err := sql.Open("mysql", conn)
	if err != nil {
		fmt.Println("Connection to mysql failed:", err)
		return
	}
	DB.SetConnMaxLifetime(100 * time.Second)
	DB.SetMaxOpenConns(100)
	sql := `show tables;`
	rows, err := DB.Query(sql)

	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(rows)
}
