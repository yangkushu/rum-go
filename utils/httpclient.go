package utils

import (
	"bytes"
	"context"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/yangkushu/rum-go/log"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"os/exec"
	"path"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

type ConfigProxy struct {
	Proxy *struct {
		UseProxy  bool   `mapstructure:"use-proxy" yaml:"use-proxy"`
		ProxyAddr string `mapstructure:"proxy-addr" yaml:"proxy-addr"`
	} `mapstructure:"proxy" yaml:"proxy"`
}

var (
	proxyUrl *url.URL = nil
)

// 签名
func Sign(params map[string]string, secretkey string) string {
	keys := make([]string, len(params))

	i := 0
	for k, _ := range params {
		keys[i] = k
		i++
	}

	sort.Strings(keys)

	values := make([]string, len(params))
	i = 0
	for _, k := range keys {
		values[i] = fmt.Sprintf("%s=%s", k, params[k])
		i++
	}

	signstr := strings.Join(values, "&") + secretkey

	return SignBuffer([]byte(signstr))
}

func SignBuffer(buff []byte) string {
	hasher := sha512.New()
	hasher.Write(buff)
	return strings.ToUpper(hex.EncodeToString(hasher.Sum(nil)))
}

func GetProxyUrl() *url.URL {
	return proxyUrl
}

func getProxyMacOS() (proxy string) {
	proxy = ""
	cmd := exec.Command("scutil", "--proxy")
	output, err := cmd.Output()
	if err != nil {
		return
	}

	lines := strings.Split(string(output), "\n")

	data := make(map[string]string)

	for _, ln := range lines {
		parts := strings.Split(ln, " : ")
		if 2 != len(parts) {
			continue
		}
		data[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}

	val, ok := data["HTTPSEnable"]
	if !ok {
		return
	}
	n, _ := strconv.Atoi(val)
	if 1 != n {
		return
	}

	addr, ok := data["HTTPSProxy"]
	if !ok {
		return
	}

	port, ok := data["HTTPSPort"]
	if !ok {
		return
	}

	proxy = fmt.Sprintf("http://%s:%s", addr, port)
	return
}

func CheckProxy(cfg *ConfigProxy) {
	if nil == cfg || nil == cfg.Proxy {
		return
	}

	// darwin, windows, linux
	proxyAddress := ""
	if runtime.GOOS == "darwin" {
		proxyAddress = getProxyMacOS()
	} else if cfg.Proxy.UseProxy {
		proxyAddress = cfg.Proxy.ProxyAddr
	}

	if len(proxyAddress) > 0 {
		u, err := url.Parse(proxyAddress)
		if IsSuccessed(err) {
			log.Info("HTTPS Proxy: " + proxyAddress)
			proxyUrl = u
			return
		}
	}

	log.Info("NO HTTPS Proxy")
}

func HttpRespond(w http.ResponseWriter, content string) {
	if n, err := w.Write([]byte(content)); nil != err {
		fmt.Println("HttpRespond fail:", err.Error())
	} else if n != len(content) {
		fmt.Println("HttpRespond fail: respond not completed")
	}
}

type RequestData struct {
	RequestString string            // 请求字符串
	Url           string            // 请求地址
	Headers       map[string]string // 请求头
	ClArgs        map[string]string
	Method        string
}

func PostJsonByData(req *RequestData) (string, error) {
	return PostJson(
		req.Url,
		req.RequestString,
		req.Headers,
		req.ClArgs,
		req.Method,
	)
}

func PostData(postUrl string, headers map[string]string, values map[string]string) ([]byte, error) {
	if nil == values || 0 == len(values) {
		return nil, errors.New("not data")
	}

	client := &http.Client{}

	if nil != proxyUrl {
		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		}
	}

	data := url.Values{}
	for key, val := range values {
		data.Set(key, val)
	}

	encodedData := data.Encode()

	req, err := http.NewRequest(http.MethodPost, postUrl, strings.NewReader(encodedData))
	if err != nil {
		return nil, err
	}

	if nil == headers {
		headers = make(map[string]string)
	}
	headers[ContentTypeHeaderName] = "application/x-www-form-urlencoded"
	headers["Content-Length"] = strconv.Itoa(len(encodedData))

	if nil != headers && len(headers) > 0 {
		for key, val := range headers {
			req.Header.Set(key, val)
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	var e error = nil

	if http.StatusOK != resp.StatusCode {
		e = errors.New(fmt.Sprintf("StatusCode[%d]", resp.StatusCode))
	}

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	return body, e
}

// PostJson 发送json数据 postUrl:请求地址 content:请求内容 headers:请求头(or nil) clArgs:请求参数(or nil)
func PostJson(postUrl, content string, headers, clArgs map[string]string, method string) (string, error) {
	if buff, err := PostJsonBytes(postUrl, content, headers, clArgs, method); IsFailed(err) {
		return string(buff), err
	} else {
		return string(buff), nil
	}
}

func PostJsonBytes(postUrl, content string, headers, clArgs map[string]string, method string) ([]byte, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	if nil != proxyUrl {
		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		}
	}

	req, err := http.NewRequest(method, postUrl, strings.NewReader(content))
	if err != nil {
		return nil, err
	}

	if nil != headers && len(headers) > 0 {
		for key, val := range headers {
			req.Header.Set(key, val)
		}
	}

	if nil != clArgs && len(clArgs) > 0 {
		q := req.URL.Query()
		for key, val := range clArgs {
			q.Add(key, val)
		}
		req.URL.RawQuery = q.Encode()
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := resp.Body.Close(); IsFailed(err) {
			fmt.Println("resp.Body.Close() fail", err.Error(), postUrl)
		}
	}()

	var e error = nil

	if 200 != resp.StatusCode {
		e = errors.New(fmt.Sprintf("StatusCode[%d]", resp.StatusCode))
	}

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	return body, e
}

// PostJsonStream 使用流方式接受回馈数据
func PostJsonStream(postUrl, content string, headers, clArgs map[string]string, method string, reader StreamReader) error {
	client := &http.Client{}

	if nil != proxyUrl {
		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		}
	}

	req, err := http.NewRequest(method, postUrl, strings.NewReader(content))
	if err != nil {
		return err
	}

	if nil != headers && len(headers) > 0 {
		for key, val := range headers {
			req.Header.Set(key, val)
		}
	}

	if nil != clArgs && len(clArgs) > 0 {
		q := req.URL.Query()
		for key, val := range clArgs {
			q.Add(key, val)
		}
		req.URL.RawQuery = q.Encode()
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	reader.BindResponse(resp)

	if 200 != resp.StatusCode {
		return errors.New(fmt.Sprintf("StatusCode[%d]", resp.StatusCode))
	}

	reader.BindResponse(resp)

	return nil
}

func DownloadFileTimeout(downUrl, fullfilename string, timeout int) error {
	if timeout < 8 {
		timeout = 8
	}
	if timeout > 60 {
		timeout = 120
	}

	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			// 使用Dialer来设定超时
			dialer := net.Dialer{
				Timeout: time.Duration(timeout) * time.Second,
			}
			// 使用Dialer的DialContext进行连接
			return dialer.DialContext(ctx, network, addr)
		},
	}
	if nil != proxyUrl {
		transport.Proxy = http.ProxyURL(proxyUrl)
	}

	client := &http.Client{
		Transport: transport,
	}

	res, err := client.Get(downUrl)
	if err != nil {
		return err
	}
	defer func() {
		if err := res.Body.Close(); nil != err {
			fmt.Println("res.Body.Close fail", err.Error())
		}
	}()

	if res.StatusCode != 200 {
		return errors.New(fmt.Sprintf("HTTP StatusCode [%d]", res.StatusCode))
	}

	f, err := os.Create(fullfilename)
	if nil != err {
		return err
	}

	_, err = io.Copy(f, res.Body)
	if e := f.Close(); nil != e {
		return e
	}

	if err != nil {
		return err
	}

	return nil
}

func DownloadFile(url, fullfilename string) error {
	res, err := http.Get(url)
	if err != nil {
		return err
	}

	defer func() {
		if err := res.Body.Close(); nil != err {
			fmt.Println("res.Body.Close fail", err.Error())
		}
	}()

	if res.StatusCode != 200 {
		return errors.New(fmt.Sprintf("HTTP StatusCode [%d]", res.StatusCode))
	}

	f, err := os.Create(fullfilename)
	if nil != err {
		return err
	}

	_, err = io.Copy(f, res.Body)
	if e := f.Close(); nil != e {
		return e
	}

	if err != nil {
		return err
	}

	return nil
}

type PostPart struct {
	Name     string
	Value    string
	Type     int    // 1 字符串；2 文件
	MimeType string // 默认空字符串，由系统决定
}

const (
	PostPartTypeString = 1
	PostPartTypeFile   = 2
)

const (
	ContentTypeHeaderName = "Content-Type"
)

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func PostMultiPart(postUrl string, headers map[string]string, parts []PostPart) ([]byte, map[string]string, error) {
	if nil == parts || 0 == len(parts) {
		return nil, nil, errors.New("no post part")
	}

	if nil != headers {
		delete(headers, ContentTypeHeaderName)
	}

	client := &http.Client{}

	if nil != proxyUrl {
		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		}
	}

	// 创建一个buffer，用来写入我们的multipart数据
	var buffer bytes.Buffer
	// 创建一个multipart writer 实例
	writer := multipart.NewWriter(&buffer)

	// 添加一个表单字段
	for _, part := range parts {
		if PostPartTypeString == part.Type {
			_ = writer.WriteField(part.Name, part.Value)
		} else if PostPartTypeFile == part.Type {
			h := make(textproto.MIMEHeader)
			h.Set("Content-Disposition",
				fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
					quoteEscaper.Replace(part.Name), quoteEscaper.Replace(path.Base(part.Value))))
			if 0 == len(part.MimeType) {
				h.Set("Content-Type", "application/octet-stream")
			} else {
				h.Set("Content-Type", part.MimeType)
			}

			//fileWriter, err := writer.CreateFormFile(part.Name, path.Base(part.Value))
			fileWriter, err := writer.CreatePart(h)
			if nil != err {
				return nil, nil, err
			}
			// 打开文件句柄
			file, err := os.Open(part.Value)
			if err != nil {
				return nil, nil, err
			}
			defer func() {
				if err = file.Close(); nil != err {
					fmt.Println("PostMultiPart close file fail", err.Error())
				}
			}()
			// 复制文件内容到fileWriter
			_, err = io.Copy(fileWriter, file)
			if err != nil {
				return nil, nil, err
			}
		}
	}

	// 关闭writer以完成multipart消息的构造
	if err := writer.Close(); nil != err {
		return nil, nil, err
	}

	// 创建一个POST请求
	req, err := http.NewRequest(http.MethodPost, postUrl, &buffer)
	if err != nil {
		return nil, nil, err
	}

	if nil != headers && len(headers) > 0 {
		for key, val := range headers {
			req.Header.Set(key, val)
		}
	}

	// 设置Content-Type头部，这很重要，因为它包含了分界符信息(boundary)
	req.Header.Set(ContentTypeHeaderName, writer.FormDataContentType())

	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}

	defer func() {
		if err := resp.Body.Close(); nil != err {
			fmt.Println("resp.Body.Close() fail",
				err.Error())
		}
	}()

	var e error = nil

	if 200 != resp.StatusCode {
		e = errors.New(fmt.Sprintf("StatusCode[%d]", resp.StatusCode))
	}

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, nil, err
	}

	headers = make(map[string]string)
	for k, v := range resp.Header {
		headers[k] = strings.Join(v, ",")
	}

	return body, headers, e
}
