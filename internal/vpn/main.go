package vpn

import (
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/sirupsen/logrus"
	"path/filepath"
	"strconv"
	"strings"
	"vpntoproxy/internal/config"
	"vpntoproxy/internal/docker"
	"vpntoproxy/internal/network"
)

// Метод создания vpn прокси-серверов
func Create(path string) (*types.Container, error) {
	logrus.Debug(">>> Starting create vpn")
	logrus.Debug("Config path: ", path)

	conf := config.Get()

	cli, err := docker.New()
	if err != nil {
		return nil, err
	}

	//pathToConfigs := conf.Proxy.PathToConfigs
	//projectDir := config.ProjectDir

	logrus.Debug(
		"MaxAttempts: ", conf.Docker.MaxAttempts,
		" StartingPort: ", conf.Proxy.StartingPort,
		//" PathToConfigs: ", pathToConfigs,
		//" ProjectDir: ", projectDir,
	)

	port := network.GetFreePort(conf.Docker.MaxAttempts, conf.Proxy.StartingPort)

	imagesList, err := cli.GetListImages()
	if err != nil {
		return nil, err
	}
	imagesListStr := strings.Trim(strings.Replace(fmt.Sprint(imagesList), " ", ",", -1), "[]")

	if !strings.Contains(imagesListStr, conf.Docker.ImageName) {
		logrus.Debug("Image not exist")
		im, err := cli.BuildImage(conf.Docker.ImageName, "")
		if err != nil {
			return nil, err
		}
		logrus.Debug("Image: ", im)
	} else {
		logrus.Debug("Image ", conf.Docker.ImageName, " exist")
	}

	_config := &container.Config{
		Image:        conf.Docker.ImageName,
		ExposedPorts: network.MakePortSet(conf.Docker.ProxyPort),
		Env: []string{
			fmt.Sprintf("PROXY_PORT=%d", conf.Docker.ProxyPort),
			fmt.Sprintf("PROXY_USER=%s", conf.Docker.ProxyUser),
			fmt.Sprintf("PROXY_PASSWORD=%s", conf.Docker.ProxyPassword),
		},
	}

	hostConfig := &container.HostConfig{
		Binds: []string{fmt.Sprintf("%s:/vpn/config.ovpn", path)},
		PortBindings: nat.PortMap{
			nat.Port(fmt.Sprintf("%d/tcp", conf.Docker.ProxyPort)): []nat.PortBinding{
				{
					HostPort: strconv.Itoa(port),
				},
			},
		},
		CapAdd: []string{"NET_ADMIN"},
		DNS:    conf.Docker.DNS,
	}

	res, err := cli.RunContainer(_config, hostConfig, filepath.Base(path))
	if err != nil {
		return nil, err
	}

	_container, err := cli.GetContainerByID(res.ID)
	if err != nil {
		return nil, err
	}

	logrus.Debug("Method «Create» completed, vpn created successfully")
	logrus.Debug("<<< Ending create vpn")

	return _container, nil
}

type TestResp struct {
	IP string `json:"origin"`
}

// Тестирование vpn проски, проверяет что внешний IP соответствует последнему прокси-серверу
/*func TestConnection(vpnAddress string, contProxy uint16, testURL string) (bool, error) {
	log.Debug(">>> Starting test proxychain")
	log.Debug("Last proxy in chain: ", chainAddress, " Container port: ", contProxy, " Testing URL: ", testURL)

	proxyUrl, err := url.Parse(fmt.Sprintf("socks5://0.0.0.0:%d", contProxy))
	http.DefaultTransport = &http.Transport{Proxy: http.ProxyURL(proxyUrl)}

	resp, err := http.Get(testURL)
	if err != nil {
		log.Debug("Failed get http request")
		return false, err
	}
	defer func(){
		err = resp.Body.Close()
		log.Error(err)
	}()

	if resp.StatusCode != 200 {
		log.Debug("Response code: ", resp.StatusCode, ", request failed")
		return false, nil
	}

	_json := TestResp{}
	err = json.NewDecoder(resp.Body).Decode(&_json)
	if err != nil {
		log.Debug("Error decode json")
		return false, err
	}

	log.Debug("<<< Ending test proxychain")

	if _json.IP == chainAddress {
		log.Debug("Proxychain passed the test successfully")
		return true, nil
	} else {
		log.Debug("Proxychain failed test")
		return false, nil
	}

}
*/
