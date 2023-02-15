package myip

import (
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"sync/atomic"
)

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

	if !ip.IsGlobalUnicast() {
		return nil, errors.New("not a global unicast address")
	}

	if ip.To4() == nil {
		return nil, errors.New("not a IPv4 address")
	}

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

	if !ip.IsGlobalUnicast() {
		return nil, errors.New("not a global unicast address")
	}
	if ip.To4() == nil {
		return nil, errors.New("not a IPv4 address")
	}

	return ip, nil
}

func getMyIpV3() (net.IP, error) {
	client := http.Client{}
	req, _ := http.NewRequest("GET", "https://api-ipv4.ip.sb/jsonip", nil)
	req.Header.Set("User-Agent", "Mozilla")

	r, err := client.Do(req)
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

	if !ip.IsGlobalUnicast() {
		return nil, errors.New("not a global unicast address")
	}
	if ip.To4() == nil {
		return nil, errors.New("not a IPv4 address")
	}

	return ip, nil
}

func getMyIpV4() (net.IP, error) {
	r, err := http.Get("https://api.ipify.org/?format=json")
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

	if !ip.IsGlobalUnicast() {
		return nil, errors.New("not a global unicast address")
	}
	if ip.To4() == nil {
		return nil, errors.New("not a IPv4 address")
	}

	return ip, nil
}

func getMyIpV5() (net.IP, error) {
	r, err := http.Get("http://myip.ipip.net/json")
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

	ipStr := res["data"].(map[string]interface{})["ip"].(string)
	ip := net.ParseIP(ipStr)

	if !ip.IsGlobalUnicast() {
		return nil, errors.New("not a global unicast address")
	}
	if ip.To4() == nil {
		return nil, errors.New("not a IPv4 address")
	}

	return ip, nil
}

func getMyIpV6() (net.IP, error) {
	r, err := http.Get("http://ip.3322.net/")
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

	s := strings.TrimSpace(string(b))
	ip := net.ParseIP(s)

	if !ip.IsGlobalUnicast() {
		return nil, errors.New("not a global unicast address")
	}
	if ip.To4() == nil {
		return nil, errors.New("not a IPv4 address")
	}

	return ip, nil
}

func GetMyIp() (net.IP, error) {

	gotIp := make(chan net.IP)
	gotErr := make(chan error)

	errCount := int32(0)
	functions := []func() (net.IP, error){getMyIp, getMyIpV2, getMyIpV3, getMyIpV4, getMyIpV5, getMyIpV6}
	for _, fun := range functions {
		fun2 := fun

		go func() {
			ip, err := fun2()
			if err != nil {
				atomic.AddInt32(&errCount, 1)
				if errCount == int32(len(functions)) {
					gotErr <- err
				}
			} else {
				gotIp <- ip
			}
		}()
	}

	select {
	case ip := <-gotIp:
		return ip, nil
	case err := <-gotErr:
		return nil, err
	}
}
