package server

import (
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/NiuStar/log"
	"github.com/NiuStar/utils"
	"github.com/NiuStar/xsql"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/net/websocket"
	"io/ioutil"
	xlog "log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/DeanThompson/ginpprof"
	"github.com/NiuStar/ftpserver"
	xgin "github.com/NiuStar/server/gin"
	"runtime"
)

type XServer struct {
	xs  *xsql.XSql
	db  *sql.DB
	db2 *sql.DB

	HadAllowAllMethod bool

	address string
	//server *http.Server
	server   *gin.Engine
	listener net.Listener

	isGraceful      bool
	signalChan      chan os.Signal
	downloadHandler func(filename, username, ip string, fullPath string)

	getRoutes    map[string]xgin.HandlerFunc
	postRoutes   map[string]xgin.HandlerFunc
	socketRoutes map[string]websocket.Handler
	config       map[string]interface{}

	w *xgin.WebSocketServices
	//s.HandfuncWebSocket("/ws",websocket.Handler(echoHandler))

}

func (s *XServer) readConfig() {
	data := utils.ReadFile("config.json")
	config1 := make(map[string]interface{})
	err := json.Unmarshal([]byte(data), &config1)
	if err != nil {
		fmt.Println("config文件错误")
		panic(err)
		return
	}

	s.config = config1
	/*s.config["DEBUG"] = config1["DEBUG"]
	s.config["port"] = config1["port"]
	s.config["tls"] = config1["tls"]
	s.config["process"] = config1["process"]
	s.config["fileName"] = config1["fileName"]
	s.config["FileList"] = config1["FileList"]*/
	/*j2 := config1["sql"].(map[string]interface{})

	j3 := config1["tlsCert"].(map[string]interface{})
	s.config["tlsCert"] = j3
	s.config["sql"] = j2

	if config1["sql2"] != nil {
		j4 := config1["sql2"].(map[string]interface{})
		s.config["sql2"] = j4
	}

	if config1["ftp"] != nil {
		s.config["ftp"] = config1["ftp"].(map[string]interface{})
	}
	*/
	ChangeTitle(config1["fileName"].(string))

	//s.config["mp4"] = config1["mp4"]
}

func Default() *XServer {

	ser := NewServer(":", DEFAULT_READ_TIMEOUT, DEFAULT_WRITE_TIMEOUT)
	ser.config = make(map[string]interface{})
	ser.getRoutes = make(map[string]xgin.HandlerFunc)
	ser.postRoutes = make(map[string]xgin.HandlerFunc)
	ser.socketRoutes = make(map[string]websocket.Handler)

	ser.w = xgin.InitSocketService()

	ser.readConfig()
	ser.address = ":" + ser.config["port"].(string)
	return ser
	//return ser
}

func (s *XServer) Config() map[string]interface{} {
	return s.config
}

func (s *XServer) GET(key string, call xgin.HandlerFunc) {

	s.w.AddGETService(key, call)
	s.getRoutes[key] = call

}
func (s *XServer) POST(key string, call xgin.HandlerFunc) {

	fmt.Println("POST方法")
	s.w.AddPOSTService(key, call)
	s.postRoutes[key] = call

}

func (s *XServer) AllMethod(call gin.HandlerFunc) {

	s.HadAllowAllMethod = true
	s.server.AllRoute(call)
}

func (s *XServer) HandfuncWebSocket(key string, call websocket.Handler) {
	s.socketRoutes[key] = call
}

func (s *XServer) InitOldSql() *sql.DB {
	j2 := s.config["sql"].(map[string]interface{})
	// db, _ = sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/suitang?charset=utf8")
	s.db, _ = sql.Open("mysql", j2["name"].(string)+":"+j2["password"].(string)+"@tcp("+j2["ip"].(string)+":"+j2["port"].(string)+")/"+j2["table"].(string)+"?charset=utf8mb4")
	s.db.SetMaxOpenConns(2000)
	s.db.SetMaxIdleConns(1000)
	s.db.Ping()
	return s.db
}

func (s *XServer) InitOldSql2() *sql.DB {
	j2 := s.config["sql2"].(map[string]interface{})
	// db, _ = sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/suitang?charset=utf8")
	s.db2, _ = sql.Open("mysql", j2["name"].(string)+":"+j2["password"].(string)+"@tcp("+j2["ip"].(string)+":"+j2["port"].(string)+")/"+j2["table"].(string)+"?charset=utf8mb4")
	s.db2.SetMaxOpenConns(2000)
	s.db2.SetMaxIdleConns(1000)
	s.db2.Ping()
	return s.db2
}

func (s *XServer) InitXSql() *xsql.XSql {
	j2 := s.config["sql"].(map[string]interface{})
	s.xs = xsql.InitSql(j2["name"].(string), j2["password"].(string), j2["ip"].(string), j2["port"].(string), j2["table"].(string))
	return s.xs
}
func (s *XServer) InitXSql3() *xsql.XSql {
	j2 := s.config["sql3"].(map[string]interface{})
	s.xs = xsql.InitSql(j2["name"].(string), j2["password"].(string), j2["ip"].(string), j2["port"].(string), j2["table"].(string))
	return s.xs
}

func (s *XServer) DownloadFileDelegate(call func(string, string, string, string)) {
	s.downloadHandler = call
}

func (this *XServer) RunServer() {

	if this.config["log_days"] != nil {
		log.SetSaveDays(int(this.config["log_days"].(float64)))
	}
	//log.SetSaveDays(int(this.config["log_days"].(float64)))

	go timer(this)

	if this.config["ftp"] != nil {
		ftp := this.config["ftp"].(map[string]interface{})
		name := ftp["name"].(string)
		pass := ftp["password"].(string)
		path := ftp["path"].(string)
		port := ftp["port"].(float64)
		ftpserver1.StartFtp(name, pass, path, int(port))
	}
	if this.config["DEBUG"].(bool) {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	if this.config["process"] != nil && this.config["process"].(bool) {
		go this.ConnectToProcess()
	}

	this.initFileServer()

	//this.server.StaticFile("/favicon.ico", utils.GetCurrPath() + "resources/favicon.ico")
	//设置静态资源
	//this.server.LoadHTMLGlob(utils.GetCurrPath() + "templates/*")

	this.server.GET("/test", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title": "Main website",
		})
	})
	this.HandfuncWebSocket("/ws", websocket.Handler(this.w.EchoHandler))
	for key, value := range this.socketRoutes {
		s := new(websocket.Server)
		s.Handler = value
		this.server.GET(key, func(c *gin.Context) {
			s.ServeHTTP(c.Writer, c.Request)
		})
	}

	for key, value := range this.getRoutes {
		this.server.GET(key, CreateGinConetxt(value))
	}
	for key, value := range this.postRoutes {

		this.server.POST(key, CreateGinConetxt(value))
	}

	fmt.Println(1)
	this.server.GET("RemoteBuild", this.remoteBuild)

	ginpprof.Wrapper(this.server)

	fmt.Println(2)
	if this.config["tls"].(bool) {
		j2 := this.config["tlsCert"].(map[string]interface{})
		this.ListenAndServerTLS(utils.GetCurrPath()+j2["cert"].(string), utils.GetCurrPath()+j2["key"].(string))
	} else {
		this.ListenAndServer()
	}
}

func CreateGinConetxt(value xgin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		xc := xgin.NewContext(c)
		value(xc)
	}

}

const (
	GRACEFUL_ENVIRON_KEY    = "IS_GRACEFUL"
	GRACEFUL_ENVIRON_STRING = GRACEFUL_ENVIRON_KEY + "=1"

	DEFAULT_READ_TIMEOUT  = 60 * time.Second
	DEFAULT_WRITE_TIMEOUT = DEFAULT_READ_TIMEOUT
)

// new server
func NewServer(addr string, readTimeout, writeTimeout time.Duration) *XServer {
	fmt.Println("port=", addr)
	log.Init()

	// 获取环境变量
	isGraceful := false
	if os.Getenv(GRACEFUL_ENVIRON_KEY) != "" {
		isGraceful = true
	}

	// 实例化Server
	return &XServer{
		address:           addr,
		server:            gin.Default(),
		isGraceful:        isGraceful,
		HadAllowAllMethod: false,
		signalChan:        make(chan os.Signal),
	}
}

func (this *XServer) ListenAndServer() error {
	return this.Server()
}

func (this *XServer) ListenAndServerTLS(certFile, keyFile string) error {

	fmt.Println("tls port =", this.address)
	config := &tls.Config{}
	srv, _ := this.server.GetTLSConfig(this.address)
	if srv.TLSConfig != nil {
		*config = *srv.TLSConfig
	}
	if config.NextProtos == nil {
		config.NextProtos = []string{"http/1.1"}
	}

	var err error
	config.Certificates = make([]tls.Certificate, 1)
	config.Certificates[0], err = tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		panic(err)
		return err
	}

	ln, err := this.server.GetTCPListener(this.isGraceful, this.address)
	//ln, err := this.server.GetTCPListener(this.isGraceful,this.address)
	if err != nil {
		panic(err)
		return err
	}
	//l := tls.NewListener(ln, config)

	this.listener = newListener(ln.(*net.TCPListener))
	return this.ServerTLS(srv, tls.NewListener(this.listener, config))
}

func (this *XServer) Server() error {

	xlog.Println(fmt.Sprintf("(this *Server) Serve() with pid %d.  "+this.address, os.Getpid()))
	ln, err := this.server.GetTCPListener(this.isGraceful, this.address)
	if err != nil {
		xlog.Println(fmt.Sprintf("1a56 with pid %d.", os.Getpid()))
		fmt.Println("1a56")
		panic(err)
		return err
	}
	this.listener = newListener(ln.(*net.TCPListener))
	// 处理信号
	go this.handleSignals()

	fmt.Println("123456")
	// 处理HTTP请求
	xlog.Println(fmt.Sprintf("123456 with pid %d.", os.Getpid()))
	this.server.Run(this.listener, this.address)

	// 跳出Serve处理代表 listener 已经close，等待所有已有的连接处理结束
	this.logf("waiting for connection close...")
	this.listener.(*Listener).Wait()
	this.logf("all connection closed, process with pid %d shutting down...", os.Getpid())

	return err
}

func (this *XServer) ServerTLS(srv http.Server, ln net.Listener) error {

	xlog.Println(fmt.Sprintf("(this *Server) Serve() with pid %d.", os.Getpid()))

	// 处理信号
	go this.handleSignals()

	fmt.Println("123456")
	// 处理HTTP请求
	xlog.Println(fmt.Sprintf("123456 with pid %d.", os.Getpid()))
	this.server.RunTLS(srv, ln)

	// 跳出Serve处理代表 listener 已经close，等待所有已有的连接处理结束
	this.logf("waiting for connection close...")
	this.listener.(*Listener).Wait()
	this.logf("all connection closed, process with pid %d shutting down...", os.Getpid())

	return nil
}

func (this *XServer) handleSignals() {
	var sig os.Signal

	signal.Notify(
		this.signalChan,
		syscall.SIGKILL,
		syscall.SIGTERM,
		syscall.SIGINT,
	)

	pid := os.Getpid()
	for {
		sig = <-this.signalChan
		/*
			switch sig {

			case syscall.SIGTERM:

				this.logf("pid %d received SIGTERM.", pid)
				this.logf("graceful shutting down http server...")

				// 关闭老进程的连接
				this.listener.(*Listener).Close()
				this.logf("listener of pid %d closed.", pid)

			case syscall.SIGINT:
				//this.StartServer()

			default:

			}*/

		fmt.Println("程序被外界关闭，信号量为：", sig)

		break
	}
	os.Exit(0)
	fmt.Println("pid:", pid)
}

func (this *XServer) StartServer() {
	pid := os.Getpid()
	fmt.Println("syscall.SIGHUP:")
	this.logf("pid %d received SIGUSR2.", pid)
	this.logf("restart http server...")

	err := this.startNewProcess()
	if err != nil {
		this.logf("start new process failed: %v, pid %d continue serve.", err)
	} else {
		// 关闭老进程的连接
		this.listener.(*Listener).Close()
		this.logf("listener of pid %d closed.", pid)
	}
}

// 启动子进程执行新程序
func (this *XServer) startNewProcess() error {

	/*listenerFd, err := this.listener.(*Listener).GetFd()
	if err != nil {
		return fmt.Errorf("failed to get socket file descriptor: %v", err)
	}*/
	listenerFd, _ := this.listener.(*Listener).GetFd()

	path := os.Args[0]

	// 设置标识优雅重启的环境变量
	environList := []string{}
	for _, value := range os.Environ() {
		if value != GRACEFUL_ENVIRON_STRING {
			environList = append(environList, value)
		}
	}
	var execSpec *syscall.ProcAttr
	environList = append(environList, GRACEFUL_ENVIRON_STRING)
	if "windows" == runtime.GOOS {
		execSpec = &syscall.ProcAttr{
			Env:   environList,
			Files: []uintptr{os.Stdin.Fd(), os.Stdout.Fd(), os.Stderr.Fd()},
		}
	} else {
		execSpec = &syscall.ProcAttr{
			Env:   environList,
			Files: []uintptr{os.Stdin.Fd(), os.Stdout.Fd(), os.Stderr.Fd(), listenerFd},
		}
	}

	fork, _, err := syscall.StartProcess(path, os.Args, execSpec)
	//fork, err := syscall.ForkExec(path, os.Args, execSpec)
	if err != nil {
		panic(err)
		return fmt.Errorf("failed to forkexec: %v", err)
	}

	this.logf("start new process success, pid %d.", fork)

	return nil
}

//获取指定目录下的所有文件，不进入下一级目录搜索，可以匹配后缀过滤。
func getNowFiles(this *XServer, dirPth string) []interface{} {
	var nowFile []interface{}
	dir, err := ioutil.ReadDir(dirPth)
	if err != nil {
		return nowFile
	}
	for _, fi := range dir {
		if fi.IsDir() { // 忽略目录
			continue
		}
		name := fi.Name()
		value := make(map[string]interface{})
		if name == this.config["fileName"].(string) {
			value["name"] = name
			value["modtime"] = fi.ModTime().Unix()
			nowFile = append(nowFile, value)
		}
	}
	//fmt.Println("getNowFiles end")
	return nowFile
}

/*
func getNowFiles(path string) []interface{} {

	var nowFile []interface{}
	//fmt.Println(path)

	fullPath := GetFullPath(path)

	//listStr := list.New()

	filepath.Walk(fullPath, func(filepath string, fi os.FileInfo, err error) error {
		if nil == fi {
			return err
		}
		if fi.IsDir() {
			return nil
		}
		name := fi.Name()
		npath := utils.Substr(filepath, len(path), len(filepath)-len(name)-len(path))
		//npath = utils.Substr(npath, 0, len(npath)-len(name))
		value := make(map[string]interface{})
		//fmt.Println(name)
		//modtime := utils.FormatTime(fi.ModTime().Unix(), "2006/01/02 - 03:04:05")
		if fi.Name() == config["fileName"].(string) {
			value["name"] = fi.Name()
			value["modtime"] = fi.ModTime().Unix()
			value["path"] = npath
			//fmt.Println(fi.Name())
			nowFile = append(nowFile, value)
		}

		return nil
	})
	return nowFile
	//OutputFilesName(listStr)
}*/

func GetFullPath(path string) string {
	absolutePath, _ := filepath.Abs(path)
	return absolutePath
}

func (this *XServer) restart() {
	fmt.Println("restart ", os.Getpid())
	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		panic(err)
		return
	}
	fmt.Println("Signal")
	//p.Kill()
	p.Signal(os.Kill)
	//p.Signal(syscall.SIGINT)
	//this.StartServer()
	go func() {
		//p.Kill()
		//p.Signal(syscall.SIGINT)
		//p.Signal(syscall.SIGTERM) //kill process
		//this.StartServer()

	}()
	//p.Kill()
	fmt.Println("Signal end ", p.Pid)
	/*pstat, err := p.Wait()
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(pstat)*/
	fmt.Println("重启服务器")
	//return
}

func timer(this *XServer) {

	var count int64 = 0
	//nt := int64(time.Now().Unix())
	timer := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-timer.C:
			{
				count++
				//fmt.Println("次数:",count)
				if len(fileList) == 0 {
					//fmt.Println("len(fileList) == 0:",count)
					fileList = getNowFiles(this, utils.GetCurrPath())
				} else {
					//fmt.Println("len(fileList) != 0",count)
					l := getNowFiles(this, utils.GetCurrPath())
					//fmt.Println(".... l := getNowFiles(\"./\") ",count)
					if len(l) != len(fileList) {
						fmt.Println("len(l) != len(fileList) : ", count)
						this.restart()
						//_ = exec.Command(fmt.Sprintf("kill -SIGUSR2 %d",os.Getpid()))

						return
					}
					for _, value := range l {
						list := value.(map[string]interface{})

						if list["modtime"].(int64) != GetTimeFromList(list["name"].(string), fileList) {
							fmt.Println("list[\"modtime\"].(int64) != GetTimeFromList(list[\"name\"].(string),fileList) :", count)
							this.restart()
							//_ = exec.Command(fmt.Sprintf("kill -SIGUSR2 %d",os.Getpid()))

							return
						}
					}
				}
			}

		}
	}
	//fmt.Println("次数:end ",count)
}

func GetTimeFromList(name string, l []interface{}) int64 {
	for _, value := range l {
		list := value.(map[string]interface{})

		if name == list["name"].(string) {
			return list["modtime"].(int64)
		}
	}
	return 0
	//list := value.(map[string]interface{})
}

func (this *XServer) logf(format string, args ...interface{}) {

	xlog.Printf(format, args...)

}
