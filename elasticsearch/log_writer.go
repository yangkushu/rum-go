package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/yangkushu/rum-go/consts"
	"github.com/yangkushu/rum-go/log"
	"github.com/yangkushu/rum-go/utils"
	"os"
	"time"
)

type LogWriter struct {
	client     *Client
	logChan    <-chan []byte
	index      string
	cancelChan chan struct{}
	// 并发度
	concurrency int
}

func NewLogWriter(client *Client, index string, logChan <-chan []byte, concurrency int) (*LogWriter, error) {
	if concurrency <= 0 {
		return nil, errors.New("concurrency must be greater than 0")
	}
	return &LogWriter{
		client:      client,
		logChan:     logChan,
		index:       index,
		concurrency: concurrency,
	}, nil
}

func (w *LogWriter) RunWrite() {
	// 从logChan中读取日志，写入es
	for i := 0; i < w.concurrency; i++ {
		go func() {
			for {
				select {
				case bs := <-w.logChan:
					// 解析日志并增加字段
					var dataMap map[string]interface{}
					var err error
					if err := json.Unmarshal(bs, &dataMap); err != nil {
						log.Error("json unmarshal error", log.ErrorField(err))
						continue
					}
					// 如果没有timestamp字段，增加一个
					if _, ok := dataMap["timestamp"]; !ok {
						dataMap["timestamp"] = time.Now().Format(consts.TimeFormatISO8601)
					}
					// ip
					if _, ok := dataMap["ip"]; !ok {
						dataMap["ip"] = utils.GetLocalIp()
					}
					// hostname
					if _, ok := dataMap["hostname"]; !ok {
						dataMap["hostname"], _ = os.Hostname()
					}
					bs, err = json.Marshal(dataMap)
					if err != nil {
						log.Error("json marshal error", log.ErrorField(err))
						continue
					}
					_, err = w.client.Index(w.getIndex()).Raw(bytes.NewReader(bs)).Do(context.Background())
					if err != nil {
						log.Error("write log to es error", log.ErrorField(err))
					}
					//log.Info("write log to es", log.String("index", w.getIndex()), log.String("response", fmt.Sprintf("%+v", resp)))
				case <-w.cancelChan:
					log.Info("log writer closed")
					return
				}
			}
		}()
	}
}

func (w *LogWriter) getIndex() string {
	// index后面加上 当前年月日，格式：ex 20210101
	return fmt.Sprintf("%s_%s", w.client.PrefixedIndex(w.index), time.Now().Format("200601"))
}

func (w *LogWriter) Close() {
	close(w.cancelChan)
}
