package webhook

import (
	"context"
	"crypto/tls"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	zlog "github.com/bind-dns/binddns-operator/pkg/utils/zaplog"
)

type AdmissionWebhookServer struct {
	HttpHandler *gin.Engine

	ListenPort      string
	ShutdownTimeout int64

	TlsCertFile string
	TlsKeyFile  string
}

func NewAdmissionWebhookServer() (server *AdmissionWebhookServer) {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	server = &AdmissionWebhookServer{
		HttpHandler: r,
	}
	server.registerRouter()
	return server
}

func (server *AdmissionWebhookServer) Run() error {
	pair, err := tls.LoadX509KeyPair(server.TlsCertFile, server.TlsKeyFile)
	if err != nil {
		zlog.Error(err)
		return err
	}

	srv := &http.Server{
		Addr:    server.ListenPort,
		Handler: server.HttpHandler,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{pair},
		},
	}
	go func(srv *http.Server, timeout int64) {
		registerGracefulShutdown(srv, timeout)
	}(srv, server.ShutdownTimeout)

	// Start
	zlog.Infof("Server serving on: %s", server.ListenPort)
	if err := srv.ListenAndServeTLS("", ""); err != nil {
		return err
	}
	return nil
}

func registerGracefulShutdown(srv *http.Server, timeout int64) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Block util a signal is received
	<-quit
	zlog.Warnf("Server is shutting down..")

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		zlog.Fatalf("Server shutdown failed, err: %s.", err.Error())
	}
	zlog.Warnf("Server exited.")
}
