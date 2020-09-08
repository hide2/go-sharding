package main

import (
	"fmt"

	"github.com/hide2/go-sharding/snowflake"
)

func main() {
	node1, err := snowflake.NewNode(1)
	if err != nil {
		fmt.Printf("Error creating NewNode, %s\n", err)
	}
	for i := 0; i < 10; i++ {
		fmt.Println("1=", node1.Generate())
	}

	node2, err := snowflake.NewNode(1000)
	if err != nil {
		fmt.Printf("Error creating NewNode, %s\n", err)
	}
	for i := 0; i < 10; i++ {
		fmt.Println("2=", node2.Generate())
	}
}
