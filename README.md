# Server

#`使用方法`

    package main
 
    import (
        "nqc.cn/server"
        "net/http"
        "nqc.cn/fmt"
        "nqc.cn/server/gin"
        "io/ioutil"
        "encoding/json"
        "strings"
    )
 
     func main() {
     
        ser := server.Default()
        ser.POST("/test1",test1)
        ser.GET("/test2",test2)
        ser.RunServer()
     }
     
    
     
     func test2(c *gin.Context) {
     
        fmt.Println("header:",c.Request.Header)
        fmt.Println("ip:",c.Request.RemoteAddr)
     
        fmt.Println("c.Request.URL:",c.Request.URL)
        id := c.DefaultQuery("id","0")
        c.String(200,id)
     }
     
     func test(c *gin.Context) {
     
        requestbody, _ := ioutil.ReadAll(c.Request.Body)
        
        c.Request.Body.Close()
     
        var list map[string]interface{}
     
        json.Unmarshal(requestbody,&list)

        var result map[string]interface{}
     
        result = make(map[string]interface{})

        fmt.Println("result:",result)
     
        result["id"] = list["id"]
        result["status"] = `{"json":45}`
        result["test"] = 2
        c.JSON(http.StatusOK,result)
     }
     
  客户端对接中可以采用HTTP/HTTPS来访问，同时可以采用websocket来访问，上述代码里已经添加了两个短链接同时生成了两个长链接接口。
  
  