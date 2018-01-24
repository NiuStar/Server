package server

import (
	"github.com/NiuStar/log/fmt"
	"github.com/NiuStar/utils"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (this *XServer) StartListenHandler() {
	var sig os.Signal

	signal.Notify(
		this.signalChan,
		/*syscall.SIGHUP,
		syscall.SIGINT,

		syscall.SIGQUIT,
		syscall.SIGILL,

		syscall.SIGTRAP,
		syscall.SIGABRT,

		syscall.SIGBUS,
		syscall.SIGFPE,
		syscall.SIGKILL,
		syscall.SIGSEGV,
		syscall.SIGPIPE,
		syscall.SIGALRM,
		syscall.SIGTERM,*/
	)

	pid := os.Getpid()

	fmt.Println("监听所有事件")
	for {
		sig = <-this.signalChan

		switch sig {

		case syscall.SIGHUP:
			fmt.Println(utils.FormatTimeAll(time.Now().Unix()), " pid= ", pid, "  syscall.SIGHUP")
		case syscall.SIGINT:
			fmt.Println(utils.FormatTimeAll(time.Now().Unix()), " pid= ", pid, "  syscall.SIGINT")
		case syscall.SIGQUIT:
			fmt.Println(utils.FormatTimeAll(time.Now().Unix()), " pid= ", pid, "  syscall.SIGQUIT")
		case syscall.SIGILL:
			fmt.Println(utils.FormatTimeAll(time.Now().Unix()), " pid= ", pid, "  syscall.SIGILL")
		case syscall.SIGTRAP:
			fmt.Println(utils.FormatTimeAll(time.Now().Unix()), " pid= ", pid, "  syscall.SIGTRAP")
		case syscall.SIGABRT:
			fmt.Println(utils.FormatTimeAll(time.Now().Unix()), " pid= ", pid, "  syscall.SIGABRT")

		case syscall.SIGBUS:
			fmt.Println(utils.FormatTimeAll(time.Now().Unix()), " pid= ", pid, "  syscall.SIGBUS")

		case syscall.SIGFPE:
			fmt.Println(utils.FormatTimeAll(time.Now().Unix()), " pid= ", pid, "  syscall.SIGFPE")
		case syscall.SIGKILL:
			fmt.Println(utils.FormatTimeAll(time.Now().Unix()), " pid= ", pid, "  syscall.SIGKILL")

		case syscall.SIGSEGV:
			fmt.Println(utils.FormatTimeAll(time.Now().Unix()), " pid= ", pid, "  syscall.SIGSEGV")
		case syscall.SIGPIPE:
			fmt.Println(utils.FormatTimeAll(time.Now().Unix()), " pid= ", pid, "  syscall.SIGPIPE")

		case syscall.SIGALRM:
			fmt.Println(utils.FormatTimeAll(time.Now().Unix()), " pid= ", pid, "  syscall.SIGALRM")
		case syscall.SIGTERM:
			fmt.Println(utils.FormatTimeAll(time.Now().Unix()), " pid= ", pid, "  syscall.SIGTERM")

		default:
			fmt.Println(utils.FormatTimeAll(time.Now().Unix()), " pid= ", pid, "  ", sig)

		}
	}
}
