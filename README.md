# 这是一个迷你的分布式系统
实现参考：https://www.bilibili.com/video/BV1ZU4y1577q

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


## 3. 注册服务
分布式系统需要有一个注册服务来管理所有接入的服务，如上述的日志服务和业务服务

具体的代码就不贴了，老样子，在registry目录下实现注册服务，然后在test目录下测试：
```
# go run main.go
```

用postman模拟请求：
```
curl --location --request POST 'http://localhost:8020/services' \
--header 'Content-Type: application/json' \
--data-raw '{
    "service_name": "LogService",
    "service_url": "localhost:8000"
}'
```

然后可以看到终端会打印出注册服务的名字和url

## 4. 服务注册
之前只是简单实现了一个注册服务，但还没有把log服务和book服务注册进去，并做到统一管理的作用，
接下来就实现这一部分。

```go
type Registration struct {
	ServiceName      ServiceName   `json:"service_name"`
	ServiceURL       string        `json:"service_url"`
	RequiredServices []ServiceName `json:"required_services"`  // 需要的其他服务
	ServiceUpdateURL string        `json:"service_update_url"` // 当前服务的客户端服务
}
```
+ 首先`Registration`结构体需要加两个字段
    + 第一个字段是当前服务依赖的其他服务，比如book服务依赖log服务
    + 第二个字段是当前服务的客户端服务，也就是注册服务给当前服务传递消息的URL
  

```go
type patchEntry struct {
    Name ServiceName
    URL  string
}

type patch struct {
    Added   []*patchEntry
    Removed []*patchEntry
}
```
+ 然后`patch`结构体是注册服务用来添加和删除服务的结构，即全局变量`reg`


接着在服务注册的函数里实现把patch发给服务的ServiceUpdateURL

