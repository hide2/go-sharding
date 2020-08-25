package main

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type ShardingConfig struct {
	Databases      []string `yaml:"databases,flow"`
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
	fmt.Println(s.TableNumber)
	fmt.Println(s.ShardingColumn)
}
