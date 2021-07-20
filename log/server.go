package log

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var logger *log.Logger

type fileLog string

func (fl fileLog) Write(data []byte) (int, error) {
	logFile, err := os.OpenFile(string(fl), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return 0, err
	}
	defer logFile.Close()
	return logFile.Write(data)
}

func InitLogger(filename string) {
	logger = log.New(fileLog(filename), "[test-log] ", log.LstdFlags | log.Lshortfile)
}

func RegisterLogHandler() {
	http.HandleFunc("/log", func(w http.ResponseWriter, r *http.Request){
		switch r.Method {
		case http.MethodPost:
			msg, err := ioutil.ReadAll(r.Body)
			if err != nil || len(msg) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			logger.Println(string(msg))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	})
}