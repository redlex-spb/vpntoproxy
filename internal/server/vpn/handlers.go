package vpn

import (
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/proxy"
	"net/http"
	"vpntoproxy/internal/config"
	"vpntoproxy/internal/docker"
	"vpntoproxy/internal/network"
	"vpntoproxy/internal/vpn"
	"vpntoproxy/pkg/requests"
	"vpntoproxy/pkg/responses"
)

// Обработка запроса на получение списка контейнеров
func list(w http.ResponseWriter, r *http.Request) {
	// TODO: middleware for log

	logrus.Debug(">>> Starting handler for get container list")

	cli, err := docker.New()
	if err != nil {
		render.JSON(w, r, responses.OutputErrorData(err))
		return
	}

	containers, err := cli.ContainersVPNList()
	if err != nil {
		render.JSON(w, r, responses.OutputErrorData(err))
		return
	}

	logrus.Debug("Containers list provided successfully")

	render.JSON(w, r, responses.OutputSuccessData(containers))

	logrus.Debug("<<< Ending handler for get container list")
}

func detail(w http.ResponseWriter, r *http.Request) {
	logrus.Debug(">>> Starting handler for get container detail")

	id := chi.URLParam(r, "ID")
	if id == "" {
		errMsg := "ID is empty"
		logrus.Error(errMsg)
		render.JSON(w, r, responses.OutputErrorData(fmt.Errorf(errMsg)))
		return
	}

	logrus.Debug("Container ID: ", id)

	cli, err := docker.New()
	if err != nil {
		render.JSON(w, r, responses.OutputErrorData(err))
		return
	}

	container, err := cli.GetContainerByID(id)
	if err != nil {
		render.JSON(w, r, responses.OutputErrorData(err))
		return
	}

	logrus.Debug("Containers detail provided successfully")

	render.JSON(w, r, responses.OutputSuccessData(container))

	logrus.Debug("<<< Ending handler for get container detail")
}

// Обработка запроса на создание vpn
func create(w http.ResponseWriter, r *http.Request) {
	logrus.Debug(">>> Starting handler for create vpn")

	body := &requests.CreateVPNParams{}
	if err := render.Bind(r, body); err != nil {
		logrus.Error(err)
		render.JSON(w, r, responses.OutputErrorData(err))
		return
	}

	logrus.Debug("Path: ", body.Path)

	info, err := vpn.Create(body.Path)
	if err != nil {
		errMsg := "Не удалось создать vpn"
		logrus.Debug("Error, cannot create config for vpn")
		logrus.Error(errMsg)
		render.JSON(w, r, responses.OutputErrorData(fmt.Errorf(errMsg)))
		return
	}

	render.JSON(w, r, responses.OutputSuccessData(info))

	logrus.Debug("Vpn created succesfully")
	logrus.Debug("<<< Ending handler for create vpn")
}

// Обработка запроса на удаление vpn
func del(w http.ResponseWriter, r *http.Request) {
	logrus.Debug(">>> Starting handler for delete vpn")

	id := chi.URLParam(r, "ID")
	if id == "" {
		errMsg := "ID is empty"
		logrus.Error(errMsg)
		render.JSON(w, r, responses.OutputErrorData(fmt.Errorf(errMsg)))
		return
	}

	logrus.Debug("Container ID: ", id)

	cli, err := docker.New()
	if err != nil {
		render.JSON(w, r, responses.OutputErrorData(err))
		return
	}

	if _, err := cli.Kill(id); err != nil {
		render.JSON(w, r, responses.OutputErrorData(err))
		return
	}
	if _, err := cli.Remove(id); err != nil {
		render.JSON(w, r, responses.OutputErrorData(err))
		return
	}

	logrus.Debug("The vpn was deleted succesfully")

	render.JSON(w, r, responses.OutputSuccessData(nil))

	logrus.Debug("<<< Ending handler for delete vpn")
}

func checkVpn(w http.ResponseWriter, r *http.Request) {
	logrus.Debug(">>> Starting handler for check vpn container")

	var id string

	q := r.URL.Query()
	if v, ok := q["id"]; !ok || v[0] == "" {
		errMsg := "ID is empty"
		logrus.Error(errMsg)
		render.JSON(w, r, responses.OutputErrorData(fmt.Errorf(errMsg)))
		return
	} else {
		id = v[0]
	}

	logrus.Debug("Container ID: ", id)

	cli, err := docker.New()
	if err != nil {
		render.JSON(w, r, responses.OutputErrorData(err))
		return
	}

	ok, err := cli.CheckVPN(id)
	if err != nil {
		render.JSON(w, r, responses.OutputErrorData(err))
		return
	}

	logrus.Debug("Vpn checked successfully")

	if ok {
		render.JSON(w, r, responses.OutputSuccessData(nil))
	} else {
		render.JSON(w, r, responses.OutputErrorData(nil))
	}

	logrus.Debug("<<< Ending handler for check vpn container")
}

func checkProxy(w http.ResponseWriter, r *http.Request) {
	logrus.Debug(">>> Starting handler for check proxy container")

	var id string

	q := r.URL.Query()
	if v, ok := q["id"]; !ok || v[0] == "" {
		errMsg := "ID is empty"
		logrus.Error(errMsg)
		render.JSON(w, r, responses.OutputErrorData(fmt.Errorf(errMsg)))
		return
	} else {
		id = v[0]
	}

	logrus.Debug("Container ID: ", id)

	cli, err := docker.New()
	if err != nil {
		render.JSON(w, r, responses.OutputErrorData(err))
		return
	}

	container, err := cli.GetContainerByID(id)
	if err != nil {
		render.JSON(w, r, responses.OutputErrorData(err))
		return
	}

	if container == nil {
		render.JSON(w, r, responses.OutputErrorData(fmt.Errorf("Container not found")))
		return
	}

	conf := config.Get()

	if container.Image != conf.Docker.ImageName {
		err = fmt.Errorf("Container image is not %s", conf.Docker.ImageName)
		render.JSON(w, r, responses.OutputErrorData(err))
		return
	}

	if len(container.Ports) < 1 {
		render.JSON(w, r, responses.OutputErrorData(fmt.Errorf("Empty container exposed ports")))
		return
	}

	proxyStr := fmt.Sprintf(
		"%s:%d",
		container.Ports[0].IP,
		container.Ports[0].PublicPort,
	)

	proxyAuth := proxy.Auth{
		User:     conf.Docker.ProxyUser,
		Password: conf.Docker.ProxyPassword,
	}

	ok, err := network.TestSuccessOfRequest(proxyStr, &proxyAuth, conf.Proxy.TestURL)
	if err != nil {
		render.JSON(w, r, responses.OutputErrorData(err))
		return
	}

	logrus.Debug("Proxy checked successfully")

	if ok {
		render.JSON(w, r, responses.OutputSuccessData(nil))
	} else {
		render.JSON(w, r, responses.OutputErrorData(nil))
	}

	logrus.Debug("<<< Ending handler for check proxy container")
}
