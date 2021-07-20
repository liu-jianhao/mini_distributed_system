package log

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
)

func NewClientLogger(url string)  {
	log.SetOutput(&clientLogger{url: url})
}

type clientLogger struct {
	url string
}

func (cl *clientLogger) Write(data []byte) (int, error) {
	fmt.Println("client logger write")

	buf := bytes.NewBuffer(data)
	resp, err := http.Post(cl.url+"/log", "text/plain", buf)
	if err != nil {
		return 0, err
	}
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("failed to send log message, response status_code=%v, status=%v", resp.StatusCode, resp.Status)
	}
	return len(data), nil
}