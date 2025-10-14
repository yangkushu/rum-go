package utils

import "net/http"

type StreamReader interface {
	BindResponse(resp *http.Response)
	Recv() (interface{}, error)
	Close()
}
