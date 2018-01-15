package server

import (
"fmt"
"os/exec"
)

func ChangeTitle(title string) {
	//执行【ls /】并输出返回文本
	_, err := exec.Command("cmd.exe", "/c", "title",title).Output()
	//f, err := exec.Command("title", "123456").Output()
	if err != nil {
		fmt.Println(err.Error())
	}

}