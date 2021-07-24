# 这是一个迷你的分布式系统

## 1. 日志系统

### 自定义log
```go
package main

import (
	"log"
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

func initLogger(filename string) {
	logger = log.New(fileLog(filename), "[test-log] ", log.LstdFlags | log.Lshortfile)
}

func main() {
	initLogger("test.log")

	logger.Println("This is a test log.")
}
```
+ 上面这个程序简单实现了一个自定义的log，这样每次用logger打印的时候都会用到自己实现的Write函数中，即写入文件中

### log server
log server的实现用到了上面的自定义log，再加一个http的handle func即可：
```go
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
```

然后简单实现一个启动http服务的函数：
```go
func StartService(ctx context.Context, hostPort string, handleFunc func()) (context.Context, error) {

	handleFunc()

	ctx, cancel := context.WithCancel(ctx)

	var server http.Server
	server.Addr = hostPort

	err := server.ListenAndServe()
	if err != nil {
		fmt.Printf("server listen and serve get err=%v\n", err)
	}
	cancel()

	return ctx, nil
}
```

最后写一个测试代码启动log server：
```go
func main() {
	log.InitLogger("test.log")

	hostPort := "localhost:8000"

	ctx, err := service.StartService(context.Background(), hostPort, log.RegisterLogHandler)
	if err != nil {
		panic(err)
	}

	<-ctx.Done()
}
```

启动服务：
```
# go run main.go
```

用postman模拟请求（可以将一下curl导入到postman中）：
```
curl --location --request POST 'http://localhost:8000/log' \
--header 'Content-Type: text/plain' \
--data-raw 'This is a test.'
```

可以看到日志文件的生成：
```
# cat test.log 
[test-log] 2021/07/20 14:31:58 server.go:36: This is a test.
[test-log] 2021/07/20 16:10:18 server.go:36: This is a test.
[test-log] 2021/07/20 16:11:55 server.go:36: This is a test.
```

## 2. 业务系统
写一个业务系统——图书管理服务

因为和分布式关系不是太大，所以只是很简单的实现。

```go
type BookHandler struct{}

func (bh *BookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	data, err := json.Marshal(bh.getAllBooks())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	_, _ = w.Write(data)
}

func (bh *BookHandler) getAllBooks() []*Book {
	return []*Book{
		{
			ID:   1,
			Name: "Go",
		},
		{
			ID:   2,
			Name: "Python",
		},
	}
}
```

然后写测试代码：
```go
func main() {
	hostPort := "localhost:8010"

	ctx, err := service.StartService(context.Background(), hostPort, book.InitBookHandler)
	if err != nil {
		panic(err)
	}

	<-ctx.Done()
}
```

运行代码：
```
# go run main.go
```

然后用postman发送请求：
```
curl --location --request GET 'http://localhost:8010/books' \
--data-raw ''
```