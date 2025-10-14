package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

func TestConnect(t *testing.T) {
	es, err := NewClient(&Config{
		Addresses: os.Getenv("address"),
		Username:  os.Getenv("username"),
		Password:  os.Getenv("password"),
	})

	if err != nil {
		panic(err)
	}

	ex, err := es.Indices.Exists("local-develop-index-test").Do(context.Background())

	if err != nil {
		panic(err)
	}

	if ex {
		fmt.Println("Index exists")
	} else {
		fmt.Println("Index does not exist")

	}

	resp, err := es.Search().Index("local-develop-index-test").Do(context.Background())

	if err != nil {
		panic(err)
	}

	fmt.Printf("result:%+v\n", resp)

	dataMap := map[string]string{
		"title":   "Test Document",
		"content": "This is a test 111",
	}

	// Convert the map to JSON
	jsonData, err := json.Marshal(dataMap)
	if err != nil {
		panic(err)
	}

	result, err := es.Index("local-develop-index-test").Id("1").Raw(bytes.NewReader(jsonData)).Do(context.Background())
	if err != nil {
		panic(err)
	}

	fmt.Printf("result:%s\n", result.Result.String())

	result2, err := es.Get("local-develop-index-test", "1").Do(context.Background())
	if err != nil {
		panic(err)
	}

	fmt.Printf("result:%+v\n", result2)
}
