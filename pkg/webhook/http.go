package webhook

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"k8s.io/api/admission/v1beta1"

	v1 "github.com/bind-dns/binddns-operator/pkg/apis/binddns/v1"
	zlog "github.com/bind-dns/binddns-operator/pkg/utils/zaplog"
)

var (
	handlerMap = make(map[string]func(*v1beta1.AdmissionReview)error)
)

func (server *AdmissionWebhookServer) registerRouter() {
	handlerMap["DnsDomain"] = server.handleDnsDomain
	handlerMap["DnsRule"] = server.handleDnsRule

	server.HttpHandler.Any("/mutate", server.mutate)
}

func (server *AdmissionWebhookServer) mutate(ctx *gin.Context) {
	ar := new(v1beta1.AdmissionReview)
	if err := ctx.BindJSON(ar); err != nil {
		zlog.Errorf("marshal request body failed, err: ", err.Error())
		ctx.JSON(http.StatusBadRequest, "marshal request body failed. err: " + err.Error())
		return
	}

	zlog.Infof("Received request. Kind: %s, Resource: %s/%s", ar.Request.Kind, ar.Request.Namespace, ar.Request.Name)

	f, ok := handlerMap[ar.Request.Kind.Kind]
	if !ok {
		zlog.Errorf("unknown kind: %s", ar.Request.Kind.Kind)
		ctx.JSON(http.StatusNotFound, "type %s handler not found")
		return
	}
	if err := f(ar); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}
	ctx.JSON(http.StatusOK, &v1beta1.AdmissionResponse{Allowed: true})
}

func (server *AdmissionWebhookServer) handleDnsDomain(ar *v1beta1.AdmissionReview) error {
	dnsDomain := new(v1.DnsDomain)
	if err := json.Unmarshal(ar.Request.Object.Raw, dnsDomain); err != nil {
		zlog.Errorf("unmarshal DnsDomain kind resource failed, err: %s", err.Error())
		return err
	}

	return nil
}

func (server *AdmissionWebhookServer) handleDnsRule(ar *v1beta1.AdmissionReview) error {
	dnsRule := new(v1.DnsRule)
	if err := json.Unmarshal(ar.Request.Object.Raw, dnsRule); err != nil {
		zlog.Errorf("unmarshal DnsRule kind resource failed, err: %s", err.Error())
		return err
	}

	return nil
}
