package server

import (
	"bufio"
	"fmt"
	"github.com/NiuStar/server/gin"
	"io"
	"net/http"
	"os/exec"
	"time"
	//"os"
)

func Build(path ,exeName string) string {

	ok, stderr := execCommand("cmd.exe", "/c", "go", "build", "-o", exeName, path)
	if ok && len(stderr) > 0 {

		return stderr
	}
	fmt.Println("state:", ok, " len: ", len(stderr))
	return ""
}

func (s *XServer) remoteBuild(c *gin.Context) {
	name := c.Query("name")

	fmt.Println("cd : " + name)
	//执行【ls /】并输出返回文本

	//ftp := s.config["ftp"].(map[string]interface{})
	//fullpath := ftp["path"].(string)
	exeName := "../" + name + "/" + name + ".exe"
	//exeName := fullpath + "/" + name + "/" + name + ".exe"
	path := "../" + name
	{

		ok, stderr := execCommand("cmd.exe", "/c", "go", "build", "-o", exeName, path)
		if ok && len(stderr) > 0 {
			c.String(http.StatusOK, stderr)
			return
		}
		fmt.Println("state:", ok, " len: ", len(stderr))
	}
	{

		execCommand("cmd.exe", "/c", "taskkill", "/f", "/im", name+".exe")
	}
	{
		go func() {
			execCommand("cmd.exe", "/c", "start", exeName)

		}()
		time.Sleep(3 * time.Second)
		/*var argv []string
		argv = append(argv, "cmd.exe")
		argv = append(argv,  "/c")
		argv = append(argv, "start")
		argv = append(argv, exeName)

		proc,err := os.StartProcess(name,argv,nil)
		_, err = proc.Wait()
		// _,stderr := execCommand("cmd.exe", "/c","start" , exeName)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("argv",argv)
		}*/
		c.String(http.StatusOK, "执行完毕")
	}
}

func execCommand(commandName string, params ...string) (bool, string) {
	cmd := exec.Command(commandName, params...)

	
	//显示运行的命令
	fmt.Println(cmd.Args)

	/*stdout, err := cmd.StdoutPipe()

	if err != nil {
		fmt.Println("err: ",err)
		return false,err.Error()
	}*/
	stderr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Println("err2", err)
		return false, err.Error()
	}

	cmd.Start()
	//fmt.Println("stdout:",getOutput(stdout))
	//fmt.Println("stderr:",getOutput(stderr))
	//go io.Copy(os.Stdout, stdout)
	//go io.Copy(os.Stderr, stderr)
	result_err := getOutput(stderr)
	cmd.Wait()
	return true, result_err
}

func getOutput(out io.ReadCloser) string {
	reader := bufio.NewReader(out)
	var result string
	//实时循环读取输出流中的一行内容
	for {
		line, ok, err2 := reader.ReadLine()
		fmt.Println("OK:", ok)
		if err2 != nil || io.EOF == err2 {
			fmt.Println("err2:", err2)
			break
		}
		result += string(line) + "\r\n"
		//fmt.Println("line:",string(line))
	}
	return result
}
