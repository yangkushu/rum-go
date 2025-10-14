package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/elastic/elastic-transport-go/v8/elastictransport"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"os"
	"strings"
)

type Client struct {
	*elasticsearch.TypedClient
	config *Config
}

func NewClient(config *Config) (*Client, error) {
	// 解析地址，为了适配之前的配置
	addrs := make([]string, 0)
	scheme := config.Scheme
	for _, addr := range strings.Split(config.Addresses, ",") {
		// 如果 addr 不以 http:// 或 https:// 开头，则添加 scheme
		if !strings.HasPrefix(addr, "http://") && !strings.HasPrefix(addr, "https://") {
			addr = scheme + "://" + addr
		}
		addrs = append(addrs, addr)
	}

	// 实现查询数据逻辑
	cfg := elasticsearch.Config{
		Addresses: addrs,
		Username:  config.Username,
		Password:  config.Password,
		//Logger:    &elastictransport.ColorLogger{Output: os.Stdout, EnableRequestBody: config.EnableRequestBody, EnableResponseBody: config.EnableRequestBody},
	}
	if config.EnableLogger {
		cfg.Logger = &elastictransport.ColorLogger{Output: os.Stdout, EnableRequestBody: config.EnableRequestBody, EnableResponseBody: config.EnableRequestBody}
	}

	client, err := elasticsearch.NewTypedClient(cfg)

	if err != nil {
		return nil, fmt.Errorf("error creating the client: %w", err)
	}
	return &Client{TypedClient: client, config: config}, nil
}

// PrefixedIndex 给索引增加前缀
func (c *Client) PrefixedIndex(index string) string {
	if c.config.IndexPrefix == "" {
		return index
	}
	// 如果索引已经有前缀，则不再添加
	if strings.HasPrefix(index, c.config.IndexPrefix) {
		return index
	}
	return c.config.IndexPrefix + index
}

// IndexStruct 插入数据
func (c *Client) IndexStruct(index string, doc interface{}) error {
	index = c.PrefixedIndex(index)
	data, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	req := esapi.IndexRequest{
		Index: index,
		Body:  strings.NewReader(string(data)),
	}
	res, err := req.Do(context.Background(), c.TypedClient)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return errors.New(fmt.Sprintf("IndexStruct failed,response :%s,", res.String()))
	}
	return nil
}

// BulkStructs 批量插入数据
func (c *Client) BulkStructs(index string, docs []interface{}) error {
	var buf bytes.Buffer
	index = c.PrefixedIndex(index)
	// 组装批量插入请求
	for _, doc := range docs {
		meta := []byte(fmt.Sprintf(`{"index":{"_index":"%s"}}%s`, index, "\n"))
		data, err := json.Marshal(doc)
		if err != nil {
			panic(err)
		}
		data = append(data, "\n"...)
		buf.Grow(len(meta) + len(data))
		buf.Write(meta)
		buf.Write(data)
	}

	// 设置批量插入请求
	req := esapi.BulkRequest{
		Index: index,
		Body:  strings.NewReader(buf.String()),
	}

	// 执行批量插入
	res, err := req.Do(context.Background(), c.TypedClient)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	// 输出结果
	if res.IsError() {
		return errors.New(fmt.Sprintf("BulkRequest failed,response :%s,", res.String()))
	}
	return nil
}

func (c *Client) Close() error {
	return c.Close()
}
