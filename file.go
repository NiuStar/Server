package server

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"nqc.cn/utils"
	"nqc.cn/fileserver"
	"nqc.cn/filestream"
	"time"

)

func (this *XServer)initFileServer() {

	var configAll []map[string]interface{}

	ok,err := utils.PathExists("host.json")
	if ok && err == nil {
		data := utils.ReadFileFullPath("host.json")

		fmt.Println("解析文件")
		err := json.Unmarshal([]byte(data), &configAll)
		if err != nil {
			fmt.Println("host文件错误")
			panic(err)
			return
		}
	}


	if this.config["FileList"] != nil {
		var list []interface{} = this.config["FileList"].([]interface{})
		for _, value := range list {
			//this.server.StaticFS("/" + value.(string), http.Dir(utils.GetCurrPath() + value.(string)))
			//this.server.StaticFSHandler("/" + value.(string), this.createStaticHandler("/" + value.(string), http.Dir(utils.GetCurrPath() + value.(string))))
			this.server.StaticFSHandler("/"+value.(string), this.createStaticHandlerOld("/"+value.(string), http.Dir(utils.GetCurrPath()+value.(string))))
		}
	} else {
		if this.HadAllowAllMethod {



			fmt.Println("this.HadAllowAllMethod")

			var folderList map[string]int = make(map[string]int)
			fmt.Println("0")

			for _,object := range configAll {
				if object["folder"] != nil {

					config := object["folder"].(map[string]interface{})

					for key,_ := range config {

						folderList[key] = 0
						//this.server.StaticFSHandler(key, this.createStaticHandler(key,configAll))
					}
				}

				if object["folder_net"] != nil {

					config := object["folder_net"].([]interface{})

					for _,value := range config {

						folderList[value.(string)] = 0
						//this.server.StaticFSHandler(value, this.createStaticHandler(value,configAll))
					}
				}
			}

			for key,_ := range folderList {
				this.server.StaticFSHandler(key, this.createStaticHandler(key,configAll))
			}
			fmt.Println("folder Over")
			//this.server.StaticFSHandler("/", this.createStaticHandler("/"))
		} else {
			fmt.Println("this.HadAllowAllMethod not")
			//this.server.StaticFS("/static",  http.Dir(utils.GetCurrPath() + "static"))
			this.server.StaticFS("/html", http.Dir(utils.GetCurrPath()+"html"))
			this.server.StaticFS("/upload", http.Dir(utils.GetCurrPath()+"upload"))
		}


	}
	fmt.Println("进入log")
	//this.server.StaticFS("/", http.Dir(utils.GetCurrPath()))
	this.server.StaticFSHandler("/log", this.createLogFileStaticHandler("/log",configAll))

	this.server.GET("/", func(c *gin.Context) {

		var proto string = "http://"
		if this.config["tls"] != nil && this.config["tls"].(bool) {
			proto = "https://"
		}

		for _,object := range configAll {

			if strings.EqualFold( object["path"].(string),c.Request.Host) {
				if object["move"] != nil  && object["move"].(bool) {

					fmt.Println("重定向",proto + c.Request.Host + "/html/index.html")
					c.Writer.Header().Add("Location",proto + c.Request.Host + "/html/index.html")

					c.Writer.WriteHeader(http.StatusMovedPermanently)
				} else if object["goto"] != nil {
					fmt.Println("重定向",proto + c.Request.Host + object["goto"].(string))
					c.Writer.Header().Add("Location",proto + c.Request.Host + object["goto"].(string))

					c.Writer.WriteHeader(http.StatusMovedPermanently)
				} else  {

					if object["folder"] != nil && object["folder"].(map[string]interface{})["html"] != nil {

						html := object["folder"].(map[string]interface{})["html"]
						fmt.Println("path:",html.(string) + "/index.html")
						c.File( html.(string) + "/index.html")

					} else {
						array := strings.Split(c.Request.URL.Path,"/")
						tempFile := utils.MD5("http://" + object["truePath"].(string) + c.Request.URL.Path)
						fmt.Println("url: ","http://" + object["truePath"].(string) + c.Request.URL.Path)

						fmt.Println( []byte(array[len(array) - 1]))

						qName := array[len(array) - 1]
						if len( qName) <= 0 {
							qName = "index.html"
						}
						fmt.Println("tempFile:",tempFile + qName)
						filedata ,type_ := utils.GETForward("http://" + object["truePath"].(string) + c.Request.URL.Path)
						err := utils.WriteToFile("./temp/" + tempFile + qName,filedata)
						if err != nil {
							panic(err)
						}
						c.Header("Content-Type",type_)
						c.File("./temp/" + tempFile + qName)
					}

				}
			}


		}

		//c.File("./html/index.html" )
	})

	fmt.Println("文件Over")
}

type fileSyatem struct {
	fs              http.FileSystem
	fileServer      http.Handler
	fileServerAdmin http.Handler
	nolisting       bool
	path            string
}

func (s *XServer) createStaticHandlerOld(relativePath string, fs http.FileSystem) gin.HandlerFunc {
	absolutePath := utils.JoinPaths(s.server.BasePath(), relativePath)
	fileServer := StripPrefix(absolutePath, fileserver.FileServer(fs))
	fileServerAdmin := StripPrefixAdmin(absolutePath, fileserver.FileServer(fs))
	//h := http.FileServer(fs)
	_, nolisting := fs.(*filestream.OnlyfilesFS)
	return func(c *gin.Context) {
		fmt.Println("header : ", c.Request)
		path := c.Request.URL.Path
		if nolisting {
			c.Writer.WriteHeader(404)
		}
		if c.Query("uid") != "nqc" {
			fileServerAdmin.ServeHTTP(c.Writer, c.Request)
		} else {
			name := c.DefaultQuery("name", "nqc")
			fileServer.ServeHTTP(c.Writer, c.Request)

			if p := strings.TrimPrefix(path, absolutePath); len(p) < len(path) {

				if strings.LastIndex(p, "/") != len(p)-1 {

					s.downloadHandler(p, name, c.Request.RemoteAddr, c.Request.RequestURI)

				}

			}
		}

	}
}

func (s *XServer) createStaticHandler(relativePath string,config []map[string]interface{}) gin.HandlerFunc {//是否为网络文件，不在本地服务器

	/*var config []map[string]string
	data := utils.ReadFileFullPath("host.json")

	err := json.Unmarshal([]byte(data), &config)
	if err != nil {
		fmt.Println("host文件错误")
		panic(err)
		return nil
	}*/

	var fsList map[string]*fileSyatem = make(map[string]*fileSyatem)
	absolutePath := utils.JoinPaths(s.server.BasePath(), relativePath)
	//fmt.Println("createStaticHandler init")

	{
		for _, value1 := range config {

			if value1["folder"] != nil {
				//fList := make(map[string]interface{})
				fList := value1["folder"].(map[string]interface{})

				for key,value := range fList {
					if strings.EqualFold(key,relativePath) {

						var fs1 http.FileSystem

						fs1 = http.Dir(utils.GetCurrPath() + value.(string))

						fmt.Println("utils.GetCurrPath() + value[\"html\"]: ", utils.GetCurrPath()+value.(string))

						fileServer := StripPrefix(absolutePath, fileserver.FileServer(fs1))

						fileServerAdmin := StripPrefixAdmin(absolutePath, fileserver.FileServer(fs1))

						_, nolisting := fs1.(*filestream.OnlyfilesFS)


						fsList[value1["path"].(string)] = &fileSyatem{fs: fs1, fileServer: fileServer, fileServerAdmin: fileServerAdmin, nolisting: nolisting, path: utils.GetCurrPath() + value.(string)}
					}
				}
			}
			//fmt.Println("1")
			if value1["folder_net"] != nil {
				//fmt.Println(value1)
				fList := value1["folder_net"].([]interface{})

				for _,value := range fList {
					fmt.Println(value)
					if strings.EqualFold(value.(string),relativePath) {
						fsList[value1["path"].(string)] = nil
					}
				}
			}
			//fmt.Println("2")
		}
	}

	value := make(map[string]string)
	//fmt.Println("进11111")
	for _, value1 := range config {
		fmt.Println("value1[\"path\"]",value1["path"])
		value[value1["path"].(string)] = value1["truePath"].(string)

	}
	//fmt.Println("进2222")

	return func(c *gin.Context) {

		array := strings.Split(c.Request.URL.Path,"/")
		fmt.Println("进入转发",c.Request.Host)
		path := c.Request.URL.Path
		if  fsList[c.Request.Host] == nil && len(value[c.Request.Host]) > 0 {
			//fmt.Println("进入转发",c.Request.Host)
			tempFile := utils.MD5(value[c.Request.Host]  + c.Request.URL.Path)

			qName := array[len(array) - 1]
			if len( qName) <= 0 {
				qName = ".html"
			}
			tm1 := time.Now()
			fmt.Println("tempFile := utils.MD5(c.Request.Host + c.Request.URL.Path):",tempFile + qName)
			filedata,type_ := utils.GETForward("http://" + value[c.Request.Host] + c.Request.URL.Path)
			err := utils.WriteToFile("./temp/" + tempFile + qName,filedata)
			if err != nil {
				panic(err)
			}
			fmt.Println("请求页面花费时长：",c.Request.URL.Path,time.Now().Sub(tm1).Seconds())
			fmt.Println("Content-Type",type_)

			c.Writer.Header().Set("Content-Type",type_)
			{
				ctypes, haveType := c.Writer.Header()["Content-Type"]
				fmt.Println("ctypes:",ctypes,"haveType:",haveType)
			}
			c.File("./temp/" + tempFile + qName)

			{
				ctypes, haveType := c.Writer.Header()["Content-Type"]
				fmt.Println("ctypes:",ctypes,"haveType:",haveType)
			}
			return
		}



		//fmt.Println("fsList[c.Request.Host].path + array[len(array) - 1]:",fsList[c.Request.Host].path + "/" + array[len(array) - 1])
		//c.File(fsList[c.Request.Host].path + "/" + array[len(array) - 1])
		//return
		fmt.Println(fsList[c.Request.Host].path ,"c.Request.URL.Path:",c.Request.URL.Path)
		nolisting := fsList[c.Request.Host].nolisting
		fileServer := fsList[c.Request.Host].fileServer
		fileServerAdmin := fsList[c.Request.Host].fileServerAdmin
		//fs := fsList[c.Request.Host].fs


		if nolisting {
			c.Writer.WriteHeader(404)
		}
		if c.Query("uid") == "nqc" {
			fileServerAdmin.ServeHTTP(c.Writer, c.Request)
		} else {
			name := c.DefaultQuery("name", "nqc")
			//fmt.Println("1 html: ",path,"   filePath: ", fsList[c.Request.Host].path)
			fileServer.ServeHTTP(c.Writer, c.Request)
			//fileServerAdmin.ServeHTTP(c.Writer, c.Request)

			if p := strings.TrimPrefix(path, absolutePath); len(p) < len(path) {
				//fmt.Println("2 html: ", strings.LastIndex(p,"/"),"   filePath: ", len(p) - 1)

				if strings.LastIndex(p, "/") != len(p)-1 {
					//fmt.Println("3 html: ",path,"   filePath: ", fsList[c.Request.Host].path)
					s.downloadHandler(p, name, c.Request.RemoteAddr, c.Request.RequestURI)

				}

			}
		}

	}
}

func (s *XServer) createLogFileStaticHandler(relativePath string,config []map[string]interface{}) gin.HandlerFunc {


	var fsList map[string]*fileSyatem = make(map[string]*fileSyatem)

	absolutePath := utils.JoinPaths(s.server.BasePath(), relativePath)

	for _, value := range config {

		if value["log"] != nil {
			var fs1 http.FileSystem
			fs1 = http.Dir(utils.GetCurrPath() + value["log"].(string))

			fmt.Println("utils.GetCurrPath() + value[\"log\"]: ", utils.GetCurrPath()+value["log"].(string))

			fileServer := StripPrefix(absolutePath, fileserver.FileServer(fs1))
			fileServerAdmin := StripPrefixAdmin(absolutePath, fileserver.FileServer(fs1))
			_, nolisting := fs1.(*filestream.OnlyfilesFS)

			fsList[value["path"].(string)] = &fileSyatem{fs: fs1, fileServer: fileServer, fileServerAdmin: fileServerAdmin, nolisting: nolisting, path: utils.GetCurrPath() + value["log"].(string) + "/../log"}
		}

		var fs1 http.FileSystem
		fs1 = http.Dir(utils.GetCurrPath() + "log")

		fmt.Println("utils.GetCurrPath() + value[\"log\"]: ", utils.GetCurrPath()+"log")

		fileServer := StripPrefix(absolutePath, fileserver.FileServer(fs1))
		fileServerAdmin := StripPrefixAdmin(absolutePath, fileserver.FileServer(fs1))
		_, nolisting := fs1.(*filestream.OnlyfilesFS)

		fsList["0"] = &fileSyatem{fs: fs1, fileServer: fileServer, fileServerAdmin: fileServerAdmin, nolisting: nolisting, path: utils.GetCurrPath() + "log"}
	}

	return func(c *gin.Context) {

		nolisting := fsList[c.Request.Host].nolisting
		//fileServer := fsList[c.Request.Host].fileServer
		fileServerAdmin := fsList[c.Request.Host].fileServerAdmin
		//fs := fsList[c.Request.Host].fs

		fmt.Println("test log")
		if c.Query("uid") == "nqc2" {
			nolisting = fsList["0"].nolisting
			//fileServer := fsList[c.Request.Host].fileServer
			fileServerAdmin = fsList["0"].fileServerAdmin
		}

		path := c.Request.URL.Path
		if nolisting {
			fmt.Println(`c.String(http.StatusNotFound,"")`)
			c.String(http.StatusNotFound, "")
		}
		if c.Query("uid") == "nqc" || c.Query("uid") == "nqc2" {
			fmt.Println(`c.Query("uid") == "nqc")`)
			fileServerAdmin.ServeHTTP(c.Writer, c.Request)
		} else {
			fmt.Println(`c.Query("uid") != "nqc")`)
			if strings.LastIndex(path, "/") == len(path)-1 {
				fmt.Println(`strings.LastIndex(path,"/")`, strings.LastIndex(path, "/"), "  len(path) - 1 = ", len(path)-1)
				c.String(http.StatusNotFound, "")

			} else {
				fmt.Println(`2 strings.LastIndex(path,"/")`, strings.LastIndex(path, "/"), "  len(path) - 1 = ", len(path)-1)
				fileServerAdmin.ServeHTTP(c.Writer, c.Request)
			}
			//c.String(http.StatusNotFound,"")
		}

	}
}

func StripPrefixAdmin(prefix string, h http.Handler) http.Handler {
	if prefix == "" {
		return h
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if p := strings.TrimPrefix(r.URL.Path, prefix); len(p) < len(r.URL.Path) {
			r.URL.Path = p
			h.ServeHTTP(w, r)

		} else {
			http.NotFound(w, r)
		}

	})
}

func StripPrefix(prefix string, h http.Handler) http.Handler {
	if prefix == "" {
		return h
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if p := strings.TrimPrefix(r.URL.Path, prefix); len(p) < len(r.URL.Path) {
			r.URL.Path = p

			if strings.LastIndex(p, "/") == len(p)-1 {
				http.NotFound(w, r)

			} else {

				h.ServeHTTP(w, r)

			}

		} else {
			http.NotFound(w, r)

		}

		//h.ServeHTTP(w, r)
	})
}

var fileList []interface{}