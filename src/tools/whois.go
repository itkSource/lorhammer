package tools

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
)

//Hostname return unique hostname_ip:port string
func Hostname(port int) (string, error) {
	name, err := os.Hostname()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s_%d", name, port), err
}

//FreeTCPPort return free tcp port on host
func FreeTCPPort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", ":0")
	if err != nil {
		return 0, err
	}
	l, err := net.ListenTCP("tcp", addr)
	defer l.Close()
	if err != nil {
		return 0, err
	}
	return l.Addr().(*net.TCPAddr).Port, nil
}

func foundIP() (string, error) {
	ifaces, err := net.Interfaces()

	if err != nil {
		return "", err
	}

	ips := make([]string, 0)

	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			splitIP := strings.Split(ip.String(), ".")

			if len(splitIP) == 4 && strings.Compare(splitIP[0], "127") != 0 && strings.Compare(splitIP[0], "172") != 0 {
				ips = append(ips, ip.String())
			}
		}
	}

	if len(ips) > 1 {
		return "", errors.New("Multiple ip found please set one")
	}
	if len(ips) == 0 {
		return "", errors.New("No ip found please set one")
	}
	return ips[0], nil
}
