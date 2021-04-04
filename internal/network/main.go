package network

import (
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/proxy"
	"net"
	"net/http"
	"strconv"
	"time"
)

// Проверка свободен ли порт
func CheckFreePort(port int) bool {
	logrus.Debug(fmt.Sprintf(">>> Starting check port :%d", port))

	portStr := strconv.Itoa(port)
	conn, err := net.DialTimeout("tcp", net.JoinHostPort("", portStr), time.Second)
	if err != nil {
		logrus.Debug("Connecting error:", err)
		return true
	}
	if conn != nil {
		defer func() {
			if err := conn.Close(); err != nil {
				logrus.Error(err)
			}
		}()
		logrus.Debug("Opened :%d", port)
		return false
	}
	return false
}

// Перебирает порты, пока порт не свободен
func GetFreePort(maxAttempts int, startingPort int) int {
	logrus.Debug(">>> Starting get free port")

	port := startingPort

	for !CheckFreePort(port) && maxAttempts > 0 {
		port++
		maxAttempts--
	}

	logrus.Debug("Selected free port: ", port)
	logrus.Debug("<<< Ending get free port")

	return port
}

// Создание объекта «PortSet», для конфигурирования контейнера при создании vpn (vpn.Create)
func MakePortSet(port int) nat.PortSet {
	logrus.Debug(">>> Starting make PortSet")

	portSet := make(nat.PortSet)

	newPort, err := nat.NewPort("tcp", strconv.Itoa(port))
	if err != nil {
		logrus.Error(err)
	}

	portSet[newPort] = struct{}{}

	logrus.Debug("<<< Ending make PortSet")

	return portSet
}

func TestSuccessOfRequest(proxyString string, proxyAuth *proxy.Auth, testUrl string) (bool, error) {
	logrus.Debug(">>> Starting test success of request")

	dialer, err := proxy.SOCKS5("tcp", proxyString, proxyAuth, proxy.Direct)
	if err != nil {
		return false, err
	}

	httpTransport := &http.Transport{}
	httpClient := &http.Client{Transport: httpTransport}
	httpTransport.Dial = dialer.Dial

	resp, err := httpClient.Get(testUrl)
	if err != nil {
		return false, err
	}

	logrus.Debug("<<< Ending test success of request")

	return resp.StatusCode == 200, nil
}
