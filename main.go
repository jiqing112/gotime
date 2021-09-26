package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"net"
	"runtime"
	"time"

	"github.com/gogf/gf/os/glog"
	"github.com/gogf/gf/os/gproc"
	"github.com/gogf/gf/text/gstr"
)

const ntpEpochOffset = 2208988800

type packet struct {
	Settings       uint8
	Stratum        uint8
	Poll           int8
	Precision      int8
	RootDelay      uint32
	RootDispersion uint32
	ReferenceID    uint32
	RefTimeSec     uint32
	RefTimeFrac    uint32
	OrigTimeSec    uint32
	OrigTimeFrac   uint32
	RxTimeSec      uint32
	RxTimeFrac     uint32
	TxTimeSec      uint32
	TxTimeFrac     uint32
}

func main() {
	var timeLayoutStr = "2006-01-02 15:04:05.000" //时间转换的模板，golang里面只能是 "2006-01-02 15:04:05" （go的诞生时间），.000是把时间精确到毫秒
	// getremotetime()返回这种格式:   2021-08-29 17:43:37.747814017 +0800 CST
	ntime := getremotetime()
	ts := ntime.Format(timeLayoutStr) //time转string
	fmt.Print(ts)

	UpdateSystemDate(ts)
}

func getremotetime() time.Time {
	var host string
	// 182.92.12.11:123 是阿里的ntp服务器，可以换成NTP Server
	flag.StringVar(&host, "e", "time.windows.com:123", "NTP host")
	flag.Parse()

	conn, err := net.Dial("udp", host)
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()
	if err := conn.SetDeadline(time.Now().Add(15 * time.Second)); err != nil {
		log.Fatalf("failed to set deadline: %v", err)
	}

	req := &packet{Settings: 0x1B}

	if err := binary.Write(conn, binary.BigEndian, req); err != nil {
		log.Fatalf("failed to send request: %v", err)
	}

	rsp := &packet{}
	if err := binary.Read(conn, binary.BigEndian, rsp); err != nil {
		log.Fatalf("failed to read server response: %v", err)
	}

	secs := float64(rsp.TxTimeSec) - ntpEpochOffset
	nanos := (int64(rsp.TxTimeFrac) * 1e9) >> 32

	showtime := time.Unix(int64(secs), nanos)

	return showtime
}

func UpdateSystemDate(dateTime string) bool {
	system := runtime.GOOS
	switch system {
	case "windows":
		{
			_, err1 := gproc.ShellExec(`date  ` + gstr.Split(dateTime, " ")[0])
			_, err2 := gproc.ShellExec(`time  ` + gstr.Split(dateTime, " ")[1])
			if err1 != nil && err2 != nil {
				// if err1 != nil {
				glog.Info("更新系统时间错误:请用管理员身份启动程序!")
				return false
			}
			return true

		}
	case "linux":
		{
			_, err1 := gproc.ShellExec(`date -s  "` + dateTime + `"`)
			if err1 != nil {
				glog.Info("更新系统时间错误:", err1.Error())
				return false
			}
			return true
		}
	case "darwin":
		{
			_, err1 := gproc.ShellExec(`date -s  "` + dateTime + `"`)
			if err1 != nil {
				glog.Info("更新系统时间错误:", err1.Error())
				return false
			}
			return true
		}
	}
	return false
}
