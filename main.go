package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/schollz/progressbar/v3"
	"io"
	"math/rand"
	"net"
	"net/http"
	"os"
	"time"
)

var token = flag.String("t", "20232023", "Token")
var listenPort = flag.Int("p", 20232, "Listen port")
var dstIpaddr = flag.String("ip", "", "Target IP address")
var delayNum = rand.Intn(4000)

func getMyIp() (net.IP, error) {
	r, err := http.Get("https://ipv4.jsonip.com/")
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = r.Body.Close()
	}()

	b, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	var res map[string]interface{}
	err = json.Unmarshal(b, &res)
	if err != nil {
		return nil, err
	}

	ipStr := res["ip"].(string)
	ip := net.ParseIP(ipStr)
	return ip, nil
}

func getMyIpV2() (net.IP, error) {
	r, err := http.Get("https://ipv4.json.myip.wtf/")
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = r.Body.Close()
	}()

	b, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	var res map[string]interface{}
	err = json.Unmarshal(b, &res)
	if err != nil {
		return nil, err
	}

	ipStr := res["YourFuckingIPAddress"].(string)
	ip := net.ParseIP(ipStr)
	return ip, nil
}

func generatePort() int {
	return rand.Intn(48128) + 1024
}

func scan(lPort int, rIp net.IP) *net.UDPAddr {
	lUdpAddr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf(":%d", lPort))
	listenConn, err := net.ListenUDP("udp4", lUdpAddr)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	defer listenConn.Close()

	var rAddr *net.UDPAddr
	recvD := make(chan struct{})
	go func() {
		for {
			buf := make([]byte, 4096)
			var n int
			n, rAddr, err = listenConn.ReadFromUDP(buf)
			if err != nil {
				fmt.Println(err)
				os.Exit(-1)
			}

			if string(buf[:n]) != token {
				continue
			}

			close(recvD)
			break
		}
	}()

	send := func(port int) {
		addr := net.UDPAddr{
			Port: port,
			IP:   rIp,
		}
		_, err = listenConn.WriteToUDP([]byte(token), &addr)
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
	}

	exit := false

	for j := 1; j <= 4096 && !exit; j++ {
		fmt.Printf("第 %d 轮探测中... (listen on %s)\n", j, lUdpAddr)

		bar := progressbar.NewOptions(65535,
			progressbar.OptionSetWidth(25),
			progressbar.OptionClearOnFinish(),
			progressbar.OptionShowCount(),
			progressbar.OptionThrottle(100*time.Millisecond),
		)
		for i := 1; i <= 65535 && !exit; i++ {
			send(i)

			_ = bar.Add(1)
			//fmt.Printf("%d%%..", i*100/65535)
			time.Sleep(time.Duration(500+delayNum) * time.Microsecond)

			select {
			case <-recvD:
				_ = bar.Close()
				exit = true
			default:
			}
		}

		//time.Sleep(time.Microsecond)
		time.Sleep(time.Duration(500+delayNum) * time.Microsecond)
	}

	select {
	case <-recvD:
	default:
		fmt.Println("探测失败.")
		os.Exit(0)
	}

	fmt.Printf("探测成功. (remote %s)\n", rAddr.String())

	rIp = rAddr.IP
	for i := 0; i < 5; i++ {
		send(rAddr.Port)
		time.Sleep(time.Millisecond)
	}
	return rAddr
}

func main() {
	//delayNum =
	flag.Parse()

	var ip net.IP
	var err error
	for {
		ip, err = getMyIp()
		if err != nil {
			fmt.Println(err)
		} else {
			break
		}

		ip, err = getMyIpV2()
		if err != nil {
			fmt.Println(err)
		}

		break
	}
	fmt.Println("你的IP地址:", ip)

	rIpStr := *dstIpaddr
	if rIpStr == "" {
		fmt.Print("请输入对方IP地址: ")
		_, err = fmt.Scanln(&rIpStr)
		if err != nil {
			fmt.Println(err)
			os.Exit(-1)
		}
	}

	//lPort := generatePort()
	lPort := *listenPort
	rAddr := scan(lPort, net.ParseIP(rIpStr))

	communicat(lPort, rAddr)
}

func communicat(lPort int, rAddr *net.UDPAddr) {
	fmt.Println("建立连接...")
	host, err := os.Hostname()
	if err != nil {
		fmt.Println(err)
	}

	lUdpAddr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf(":%d", lPort))
	listenConn, err := net.ListenUDP("udp4", lUdpAddr)
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	defer listenConn.Close()

	go func() {
		for {
			buf := make([]byte, 4096)
			var n int
			n, _, err = listenConn.ReadFromUDP(buf)
			if err != nil {
				fmt.Println(err)
				os.Exit(-1)
			}
			fmt.Printf("%s Recv: %s\n", timeStr(), buf[:n])
		}
	}()

	for {
		send := "test-" + host
		fmt.Printf("%s Send: %s\n", timeStr(), send)
		_, err = listenConn.WriteToUDP([]byte(send), rAddr)
		time.Sleep(2 * time.Second)
	}
}

func timeStr() string {
	currentTime := time.Now()
	return fmt.Sprintf("%02d:%02d:%02d.%03d",
		currentTime.Hour(),
		currentTime.Hour(),
		currentTime.Second(),
		currentTime.Nanosecond()/1000_000)
}
