package router

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"

	zlog "github.com/bind-dns/binddns-operator/pkg/utils/zaplog"
)

type HttpServer struct {
	r    *gin.Engine
	port string
	srv  *http.Server
}

func NewHttpServer(port string) (server *HttpServer) {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	gin.Recovery()

	// Set pprof.
	pprof.Register(r, "/debug/pprof")

	// Set zip.
	r.Use(gzip.Gzip(gzip.DefaultCompression))

	server = &HttpServer{
		r:    r,
		port: port,
	}

	// Register http router.
	server.registerRouter()

	return server
}

func (server *HttpServer) registerRouter() {
	r := server.r

	// Static assets.
	r.StaticFS("/static/", http.Dir("webapp/assets/"))

	// Pages
	console := r.Group("/console")
	console.StaticFile("/domains", "webapp/html/domain.html")
	console.StaticFile("/rules", "webapp/html/rule.html")

	api := r.Group("/api")
	// DnsDomain CRUD.
	api.GET("/domains", listDomains)
	api.POST("/domains", createDomain)
	api.DELETE("/domains/:domain", deleteDomain)
	api.PUT("/domains/:domain/pause", pauseDomain)
	api.PUT("/domains/:domain/open", openDomain)

	// DnsRule CRUD.
	api.GET("/rules", listRules)
	api.POST("/rules", createRule)
	api.DELETE("/rules/:rule", deleteRule)
	api.PUT("/rules/:rule", updateRule)
	api.PUT("/rules/:rule/pause", pauseRule)
	api.PUT("/rules/:rule/open", openRule)
}

// Start used to start the api http server.
func (server *HttpServer) Start() {
	server.srv = &http.Server{
		Addr:    server.port,
		Handler: server.r,
	}
	zlog.Infof("HttpServer started, serving on: %s", server.port)
	if err := server.srv.ListenAndServe(); err != nil {
		panic(err)
	}
}

// Stop used to stop the http server.
func (server *HttpServer) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(20)*time.Second)
	defer cancel()
	if err := server.srv.Shutdown(ctx); err != nil {
		zlog.Fatalf("HttpServer shutdown failed, err: %s", err.Error())
	}
	zlog.Warnf("HttpServer exited.")
}
