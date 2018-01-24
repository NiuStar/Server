package gin

import (
	"encoding/json"
	"github.com/NiuStar/log/fmt"
	"github.com/NiuStar/utils"
	"golang.org/x/net/websocket"

	"bytes"
	"io/ioutil"
	//"golang.org/x/net/context"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

type H map[string]interface{}

const ORDER = "Order"
const TAG = "TAG"
const GETPARAMS = "GetParams"
const POSTPARAMS = "PostParams"
const ContentType = "Content-Type"
const ApplicationJson = "application/json"
const ApplicationUrlencoded = "application/x-www-form-urlencoded"
const STATUS = "Status"
const RESULT = "Result"
const EXT = "Ext"
const HEART = "heart"
const OTHER = "other"

type WebSocketServices struct {
	//10月25日添加短链接转长连接接口
	socketService map[string]*Method
}

//10月25日添加短链接转长连接接口

type Context struct {
	*gin.Context
	key string
	tag string
	ws  *websocket.Conn
}

type Method struct {
	mtype uint8 //类型 0 GET 1 POST
	key   string
	value HandlerFunc
}
type HandlerFunc func(*Context)

// String writes the given string into the response body.
func (c *Context) String(code int, format string) {

	if c.ws == nil {
		fmt.Println("c.ws == nil")
		//c.Context.Header()
		c.Context.String(code, format)
	} else {

		var obj interface{}

		err := json.Unmarshal([]byte(format), &obj)

		var result map[string]interface{} = make(map[string]interface{})
		result[STATUS] = true
		result[ORDER] = c.key
		result[TAG] = c.tag

		if err != nil {
			result[RESULT] = format
		} else {
			result[RESULT] = obj
		}

		result[EXT] = `成功`

		body, err := json.Marshal(result)
		if err != nil {
			fmt.Println("长连接："+c.key+c.ws.Request().RemoteAddr, err)
		}

		if err := websocket.Message.Send(c.ws, string(body)); err != nil {
			fmt.Println("长连接："+c.key+c.ws.Request().RemoteAddr, err)
		}
	}

	fmt.Println("StringOver")

}

func (c *Context) JSON(code int, obj interface{}) {

	if c.ws != nil {

		var result map[string]interface{} = make(map[string]interface{})
		result[STATUS] = true
		result[ORDER] = c.key

		result[TAG] = c.tag
		result[RESULT] = obj

		result[EXT] = `成功`

		body, err := json.Marshal(result)
		if err != nil {
			fmt.Println("长连接："+c.key+c.ws.Request().RemoteAddr, err)
		}

		if err := websocket.Message.Send(c.ws, string(body)); err != nil {
			fmt.Println("长连接："+c.key+c.ws.Request().RemoteAddr, err)
		}
	} else {
		c.Context.JSON(code, obj)
		//oc := c.gin.Context)

	}
}

func (c *Context) JSON2(code int, obj interface{}) {

	if c.ws != nil {

		var result map[string]interface{} = make(map[string]interface{})
		result[STATUS] = true
		result[ORDER] = c.key

		result[RESULT] = obj

		result[TAG] = c.tag
		result[EXT] = `成功`

		body, err := json.Marshal(result)
		if err != nil {
			fmt.Println("长连接："+c.key+c.ws.Request().RemoteAddr, err)
		}

		if err := websocket.Message.Send(c.ws, strings.Replace(string(body), "null", "[]", -1)); err != nil {
			fmt.Println("长连接："+c.key+c.ws.Request().RemoteAddr, err)
		}
	} else {
		c.Context.JSON2(code, obj)
		//oc := c.gin.Context)

	}
}

func InitSocketService() *WebSocketServices {

	w := &WebSocketServices{socketService: make(map[string]*Method)}

	return w
}

func newContext(c *gin.Context, key string, ws *websocket.Conn) *Context {
	return &Context{
		Context: c,
		key:     key,
		ws:      ws,
		//waitGroup:   &sync.WaitGroup{},
	}
}

func NewContext(c *gin.Context) *Context {
	return &Context{
		Context: c,
		key:     "",
		ws:      nil,
		//waitGroup:   &sync.WaitGroup{},
	}
}

func (w *WebSocketServices) EchoHandler(ws *websocket.Conn) {

	defer fmt.Over()
	for {
		fmt.Start()

		var reply string
		var err error
		if err = websocket.Message.Receive(ws, &reply); err != nil {
			fmt.Println("Can't receive")
			return
		}

		fmt.Println("reply:", reply)

		var result map[string]interface{} = make(map[string]interface{})

		var order map[string]interface{} = make(map[string]interface{})

		err = json.Unmarshal([]byte(reply), &order)

		fmt.Println("ORDER:", ORDER, order)

		if err != nil {
			result[STATUS] = false
			result[ORDER] = OTHER
			result[EXT] = err.Error()
		} else {

			for key, value := range order {
				order[strings.TrimSpace(key)] = value
			}

			if order[ORDER] != nil {
				key := order[ORDER].(string)
				fmt.Println("key:", key)
				if key == HEART {
					result[STATUS] = true
					result[ORDER] = "heart"
					result[EXT] = `心跳包`
				} else if w.socketService[key] == nil {
					result[STATUS] = false
					result[ORDER] = "other"
					result[EXT] = `1111没有` + key + `指令`
				} else {

					oc := &gin.Context{}

					if w.socketService[key].mtype == 1 && order[POSTPARAMS] != nil {

						bf := bytes.NewBuffer([]byte(order[POSTPARAMS].(string)))
						body := ioutil.NopCloser(bf)
						//body := ioutil.NopCloser(strings.NewReader(req1.PostForm.Encode())) //把form数据编下码

						request, _ := http.NewRequest("POST", "http://127.0.0.1/"+key+"?"+order[GETPARAMS].(string), body)

						if order["Content-Type"] != nil {
							request.Header.Add(ContentType, order[ContentType].(string))
						} else {
							request.Header.Add(ContentType, ApplicationUrlencoded)
						}

						fmt.Println("params:", order[POSTPARAMS].(string))

						oc.Request = request

					} else if w.socketService[key].mtype == 0 && order[GETPARAMS] != nil {
						request, _ := http.NewRequest("GET", "http://127.0.0.1/"+key+"?"+order[GETPARAMS].(string), nil)

						oc.Request = request
					} else {
						oc.Request = ws.Request()
					}

					//var c interface{}
					//oc := &gin.Context{Request:ws.Request()}

					c := newContext(oc, key, ws)

					if order[TAG] != nil {
						c.tag = order[TAG].(string)
					} else {
						c.tag = ""
					}

					//c := &Context{Context:oc,key:key,ws:ws}
					//c := Context{Context:oc}
					fmt.Println("调用相关方法")
					fmt.Println("c.Context:", c.Context)

					w.socketService[key].value(c)
					fmt.Println("调用相关方法结束")
					result[STATUS] = true
					result[ORDER] = OTHER
					result[EXT] = `指令发送成功`
				}
			} else {
				result[STATUS] = false
				result[ORDER] = OTHER
				result[EXT] = `没有添加指令`
			}

		}

		result[TAG] = order[TAG]
		body, err := json.Marshal(result)
		if err != nil {
			fmt.Println("长连接："+ws.Request().RemoteAddr, err)
		}

		if err := websocket.Message.Send(ws, string(body)); err != nil {
			//return
		}
		//return
	}
}

func (w *WebSocketServices) AddGETService(key string, call HandlerFunc) {

	fmt.Println("go in AddService ")
	var k string = key
	if len(key) > 1 {
		if key[0] == '/' {
			k = utils.Substr(key, 1, len(key))
		}
	}

	fmt.Println("AddService : ", k, call)

	m := &Method{mtype: 0, key: key, value: call}
	w.socketService[k] = m
}

func (w *WebSocketServices) AddPOSTService(key string, call HandlerFunc) {

	fmt.Println("go in AddService ")
	var k string = key
	if len(key) > 1 {
		if key[0] == '/' {
			k = utils.Substr(key, 1, len(key))
		}
	}

	fmt.Println("AddService : ", k, call)
	m := &Method{mtype: 1, key: key, value: call}
	w.socketService[k] = m
}
