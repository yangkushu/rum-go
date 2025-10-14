package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"
)

var (
	BuildTime = "MacVersion"
)

func GetBuildTime() string {
	// return C.GoString(C.build_time())
	return BuildTime
}

func IsSuccessed(err error) bool {
	return (nil == err)
}

func IsFailed(err error) bool {
	return (nil != err)
}

func GetNanoRandomString(l int) string {
	if str, err := gonanoid.Generate("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ", l); nil != err {
		return ""
	} else {
		return str
	}
}

func GetTimestampMs() int64 {
	return time.Now().UnixNano() / 1e6
}

func GetTimeString() string {
	return time.Now().String()
}

// GetCurrentDirectory get module directory
func GetCurrentDirectory() (string, error) {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return "", err
	}
	return strings.Replace(dir, "\\", "/", -1), nil
}

func ChangeCurrentDir() error {
	if currentDir, err := GetCurrentDirectory(); nil != err {
		fmt.Printf("get current directory fail, %v\n", err)
		return err
	} else if err := os.Chdir(currentDir); nil != err {
		fmt.Printf("change current directory fail, %v\n", err)
		return err
	} else {
		//fmt.Printf("current directory: %s\n", currentDir)
		return nil
	}
}

func Decrypt(encrypt string, key string) (string, error) {
	buf, err := base64.StdEncoding.DecodeString(encrypt)
	if err != nil {
		return "", fmt.Errorf("base64StdEncode Error[%v]", err)
	}
	if len(buf) < aes.BlockSize {
		return "", errors.New("cipher  too short")
	}
	keyBs := sha256.Sum256([]byte(key))
	block, err := aes.NewCipher(keyBs[:sha256.Size])
	if err != nil {
		return "", fmt.Errorf("AESNewCipher Error[%v]", err)
	}
	iv := buf[:aes.BlockSize]
	buf = buf[aes.BlockSize:]
	// CBC mode always works in whole blocks.
	if len(buf)%aes.BlockSize != 0 {
		return "", errors.New("ciphertext is not a multiple of the block size")
	}
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(buf, buf)
	n := strings.Index(string(buf), "{")
	if n == -1 {
		n = 0
	}
	m := strings.LastIndex(string(buf), "}")
	if m == -1 {
		m = len(buf) - 1
	}
	if n >= m {
		return "", errors.New(fmt.Sprintf("slice bounds out of range %d:%d", n, m))
	} else {
		return string(buf[n : m+1]), nil
	}
}

// 安全JSON编码
func JsonEncode(content interface{}, indent, escape bool) (string, error) {
	byteBuf := bytes.NewBuffer(make([]byte, 0))
	encoder := json.NewEncoder(byteBuf)
	if indent {
		encoder.SetIndent("", "  ")
	}
	encoder.SetEscapeHTML(escape)
	if err := encoder.Encode(content); IsSuccessed(err) {
		return byteBuf.String(), nil
	} else {
		return "", err
	}
}

func comma(s string) string {
	n := len(s)
	if n <= 3 {
		return s
	}
	return comma(s[:n-3]) + "," + s[n-3:]
}

// AddCommas 给格式化数字添加逗号分割
func AddCommas(n int) string {
	return comma(fmt.Sprintf("%d", n))
}

// 返回距离证书过期剩余的天数
// webSite 网站地址
func checkCertificate(webSiteUrl string) (days int, err error) {
	days = 0
	err = nil

	// 目标网站
	host := webSiteUrl + ":443"

	// 建立一个TLS连接以获取证书
	conn, err := tls.Dial("tcp", host, nil)
	if err != nil {
		return
	}
	defer func() {
		_ = conn.Close()
	}()

	// 获取连接状态，其中包含服务器的证书链
	state := conn.ConnectionState()
	if len(state.PeerCertificates) == 0 {
		err = errors.New("no Certificate")
		return
	}

	days = int((state.PeerCertificates[0].NotAfter.Unix() - time.Now().Unix()) / 86400)
	return
}

// OutputObjectInfo 注意不要是指针类型
func OutputObjectInfo(obj interface{}) {
	// 获取输入变量的反射类型对象
	t := reflect.TypeOf(obj)

	// 获取输入变量的反射值对象
	v := reflect.ValueOf(obj)

	// 判断输入是否为结构体类型
	if t.Kind() == reflect.Struct {
		// 获取结构体的字段数量
		num := t.NumField()
		fmt.Println("")
		for i := 0; i < num; i++ {
			//获取每个字段
			field := t.Field(i)
			//打印字段名和字段值
			fmt.Printf("   '%s' = ['%v']\n", field.Name, v.Field(i))
		}
	} else {
		fmt.Println("The input is not a struct type")
	}
}

// TrimSqlStr 去掉sql中的换行和多余的空格
func TrimSqlStr(sql string) string {
	sql = strings.ReplaceAll(sql, "\n", " ")
	sql = strings.ReplaceAll(sql, "\t", " ")
	sql = strings.ReplaceAll(sql, "  ", " ")
	return sql
}
