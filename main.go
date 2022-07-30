package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	fmt.Println("hello world")
	data, err := os.ReadFile("/config/schema.json")
	if err != nil {
		fmt.Println("error reading /config/schema.json")
	}

	schema := schema{}
	json.Unmarshal([]byte(data), &schema)

	fmt.Println(schema)
}

type schema struct {
	Databases []string `json:"databases"`
}
