package objectstorage

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"io"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

var (
	regexIPv4Address = regexp.MustCompile(`^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$`)
	forcePathStyle   = false
	// ForcePathStyle = true ：bucket 拼在路径里
	// ForcePathStyle = false ：bucket 拼在域名前
)

func isValidIP(ip string) bool {
	return regexIPv4Address.MatchString(ip)
}

func selfDialContext(ctx context.Context, network, address string) (net.Conn, error) {
	if !forcePathStyle && network == "tcp" {
		host, port, err := net.SplitHostPort(address)
		if err != nil {
			return nil, err
		}
		secs := strings.Split(host, ".")
		if len(secs) >= 4 {
			str := strings.Join(secs[len(secs)-4:], ".")
			if isValidIP(str) {
				address = str + ":" + port
			}
		}
	}
	return (&net.Dialer{}).DialContext(ctx, network, address)
}

func createCustomHTTPClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
}

type S3Client struct {
	S3     *s3.S3
	config *S3Config
}

// NewS3Client 新的构造函数接收S3Config和bucketName
func NewS3Client(c *S3Config) (IObjectStorage, error) {

	config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(c.AccessKeyID, c.SecretAccessKey, ""),
		Endpoint:         aws.String(c.Endpoint),
		Region:           aws.String(c.Region),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(c.ForcePathStyle),
	}
	sess := session.Must(session.NewSession(config))
	s3Client := s3.New(sess)

	forcePathStyle = c.ForcePathStyle

	return &S3Client{
		S3:     s3Client,
		config: c,
	}, nil
}

// UploadFile 方法现在不需要bucketName作为参数
func (c *S3Client) UploadFile(key, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer func() {
		_ = file.Close()
	}()

	_, err = c.S3.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(c.config.Bucket),
		Key:    aws.String(key),
		Body:   file,
	})
	if err != nil {
		return fmt.Errorf("failed to upload file to S3: %v", err)
	}

	return nil
}

//// DownloadFile 方法现在不需要bucketName作为参数
//func (c *S3Client) DownloadFile(filename string) ([]byte, error) {
//	resp, err := c.S3.GetObject(&s3.GetObjectInput{
//		Bucket: aws.String(c.config.Bucket),
//		Key:    aws.String(filename),
//	})
//	if err != nil {
//		return nil, fmt.Errorf("failed to download file: %v", err)
//	}
//	defer resp.Body.Close()
//
//	buf := bytes.NewBuffer(nil)
//	if _, err := io.Copy(buf, resp.Body); err != nil {
//		return nil, fmt.Errorf("failed to read file body: %v", err)
//	}
//
//	return buf.Bytes(), nil
//}

func (c *S3Client) DownloadFile(key string, filePath string) error {
	resp, err := c.S3.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(c.config.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to download file: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer func() {
		_ = file.Close()
	}()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	return nil

}

// 判断文件是否存在
func (c *S3Client) IsFileExist(key string) (bool, error) {
	_, err := c.S3.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(c.config.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var aerr awserr.Error
		if errors.As(err, &aerr) && aerr.Code() == s3.ErrCodeNoSuchKey {
			return false, nil
		}
		return false, fmt.Errorf("failed to check file existence: %v", err)
	}

	return true, nil
}

//// DeleteFile 方法现在不需要bucketName作为参数
//func (c *S3Client) DeleteObject(filename string) error {
//	_, err := c.S3.DeleteObject(&s3.DeleteObjectInput{
//		Bucket: aws.String(c.config.Bucket),
//		Key:    aws.String(filename),
//	})
//	if err != nil {
//		return fmt.Errorf("failed to delete file: %v", err)
//	}
//	return nil
//}

//func (c *S3Client) ListFiles() ([]string, error) {
//	resp, err := c.S3.ListObjects(&s3.ListObjectsInput{
//		Bucket: aws.String(c.config.Bucket),
//	})
//	if err != nil {
//		return nil, fmt.Errorf("failed to list files: %v", err)
//	}
//
//	var files []string
//	for _, item := range resp.Contents {
//		files = append(files, *item.Key)
//	}
//	return files, nil
//}
