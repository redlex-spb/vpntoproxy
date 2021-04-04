package server

import (
	"context"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

type HttpServer struct {
	server *http.Server
}

func New(port int) *HttpServer {
	logrus.Infof("Starting listening on port: %d", port)

	return &HttpServer{
		server: &http.Server{
			Addr:    ":" + strconv.Itoa(port),
			Handler: route(),
		},
	}
}

func (hs *HttpServer) Run() {
	err := hs.server.ListenAndServe()
	if err != nil {
		logrus.Info(err)
	}
}

func (hs *HttpServer) GracefulShutdown() {

	stop := make(chan os.Signal, 1)

	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	sig := <-stop

	logrus.Println("Shutting down server... Reason:", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := hs.server.Shutdown(ctx); err != nil {
		logrus.Fatalf("Error: %v\n", err)
	} else {
		logrus.Println("Server stopped gracefully")
	}
}
