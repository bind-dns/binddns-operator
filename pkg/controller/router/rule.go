package router

import (
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	binddnsv1 "github.com/bind-dns/binddns-operator/pkg/apis/binddns/v1"
	"github.com/bind-dns/binddns-operator/pkg/kube"
	"github.com/bind-dns/binddns-operator/pkg/utils"
	zlog "github.com/bind-dns/binddns-operator/pkg/utils/zaplog"
)

type DnsRuleRequest struct {
	Name       string `json:"name"`
	Zone       string `json:"zone"`
	Host       string `json:"host"`
	Type       string `json:"type"`
	Data       string `json:"data"`
	Ttl        int32  `json:"ttl"`
	MxPriority int32  `json:"mxPriority"`
}

type DnsRuleEntity struct {
	Name       string
	Zone       string
	Host       string
	Type       string
	Data       string
	Ttl        int32
	MxPriority int32
	CreateTime string
	Enabled    bool
}

func listRules(ctx *gin.Context) {
	domain := strings.TrimSpace(ctx.Query("domain"))
	if domain == "" {
		ctx.JSON(200, &Response{
			Code: ERROR,
			Msg:  "query param: domain, cannot be empty.",
		})
	}

	list, err := kube.GetKubeClient().GetDnsClientSet().BinddnsV1().DnsRules().List(ctx, v1.ListOptions{
		ResourceVersion: "0",
		LabelSelector: strings.Join([]string{
			utils.LabelZoneDnsRule + "=" + domain,
		}, ","),
	})
	if err != nil {
		zlog.Error(err)
		ctx.JSON(200, &Response{
			Code: ERROR,
			Msg:  err.Error(),
		})
		return
	}

	s := make([]*DnsRuleEntity, 0, len(list.Items))
	for i := range list.Items {
		item := &list.Items[i]
		s = append(s, &DnsRuleEntity{
			Name:       item.Name,
			Zone:       item.Spec.Zone,
			Host:       item.Spec.Host,
			Type:       string(item.Spec.Type),
			Data:       item.Spec.Data,
			Ttl:        item.Spec.Ttl,
			MxPriority: item.Spec.MxPriority,
			CreateTime: item.Status.CreateTime,
			Enabled:    item.Spec.Enabled,
		})
	}

	// Sort by createTime
	sort.Sort(DnsRuleSort(s))

	ctx.JSON(200, &Response{
		Code: SUCCESS,
		Data: s,
	})
}

func createRule(ctx *gin.Context) {
	req := new(DnsRuleRequest)
	if err := ctx.BindJSON(req); err != nil {
		zlog.Error(err)
		ctx.JSON(http.StatusOK, &Response{
			Code: ERROR,
			Msg:  err.Error(),
		})
		return
	}
	_, err := kube.GetKubeClient().GetDnsClientSet().BinddnsV1().DnsRules().Create(ctx, &binddnsv1.DnsRule{
		TypeMeta: v1.TypeMeta{
			Kind:       "DnsRule",
			APIVersion: "binddns.github.com/v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: "nothing",
		},
		Spec: binddnsv1.DnsRuleSpec{
			Zone:       req.Zone,
			Enabled:    true,
			Host:       req.Host,
			Type:       binddnsv1.DnsType(req.Type),
			Data:       req.Data,
			Ttl:        req.Ttl,
			MxPriority: req.MxPriority,
		},
	}, v1.CreateOptions{})
	if err != nil {
		ctx.JSON(200, &Response{
			Code: ERROR,
			Msg:  err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, &Response{Code: SUCCESS})
}

func deleteRule(ctx *gin.Context) {
	rule := ctx.Param("rule")
	err := kube.GetKubeClient().GetDnsClientSet().BinddnsV1().DnsRules().Delete(ctx, rule, v1.DeleteOptions{})
	if err != nil && !k8sErrors.IsNotFound(err) {
		zlog.Error(err)
		ctx.JSON(200, &Response{
			Code: ERROR,
			Msg:  err.Error(),
		})
		return
	}

	ctx.JSON(200, &Response{
		Code: SUCCESS,
		Data: nil,
	})
}

func updateRule(ctx *gin.Context) {
	req := new(DnsRuleRequest)
	if err := ctx.BindJSON(req); err != nil {
		zlog.Error(err)
		ctx.JSON(http.StatusOK, &Response{
			Code: ERROR,
			Msg:  err.Error(),
		})
		return
	}

	ruleName := ctx.Param("rule")
	dnsRule, err := kube.GetKubeClient().GetDnsClientSet().BinddnsV1().DnsRules().Get(ctx, ruleName, v1.GetOptions{})
	if err != nil {
		zlog.Error(err)
		ctx.JSON(http.StatusOK, &Response{
			Code: ERROR,
			Msg:  err.Error(),
		})
		return
	}
	dnsRule.Spec.Host = req.Host
	dnsRule.Spec.Type = binddnsv1.DnsType(req.Type)
	dnsRule.Spec.Ttl = req.Ttl
	dnsRule.Spec.Data = req.Data
	dnsRule.Spec.MxPriority = req.MxPriority
	_, err = kube.GetKubeClient().GetDnsClientSet().BinddnsV1().DnsRules().Update(ctx, dnsRule, v1.UpdateOptions{})
	if err != nil {
		zlog.Error(err)
		ctx.JSON(http.StatusOK, &Response{
			Code: ERROR,
			Msg:  err.Error(),
		})
		return
	}
	ctx.JSON(http.StatusOK, &Response{Code: SUCCESS})
}

func pauseRule(ctx *gin.Context) {
	rule := ctx.Param("rule")

	dnsRule, err := kube.GetKubeClient().GetDnsClientSet().BinddnsV1().DnsRules().Get(ctx, rule, v1.GetOptions{})
	if err != nil {
		zlog.Error(err)
		ctx.JSON(200, &Response{
			Code: ERROR,
			Msg:  err.Error(),
		})
		return
	}
	if !dnsRule.Spec.Enabled {
		ctx.JSON(200, &Response{Code: SUCCESS})
		return
	}

	dnsRule.Spec.Enabled = false
	_, err = kube.GetKubeClient().GetDnsClientSet().BinddnsV1().DnsRules().Update(ctx, dnsRule, v1.UpdateOptions{})
	if err != nil {
		zlog.Error(err)
		ctx.JSON(200, &Response{
			Code: ERROR,
			Msg:  err.Error(),
		})
		return
	}
	ctx.JSON(200, &Response{Code: SUCCESS})
}

func openRule(ctx *gin.Context) {
	rule := ctx.Param("rule")

	dnsRule, err := kube.GetKubeClient().GetDnsClientSet().BinddnsV1().DnsRules().Get(ctx, rule, v1.GetOptions{})
	if err != nil {
		zlog.Error(err)
		ctx.JSON(200, &Response{
			Code: ERROR,
			Msg:  err.Error(),
		})
		return
	}
	if dnsRule.Spec.Enabled {
		ctx.JSON(200, &Response{Code: SUCCESS})
		return
	}

	dnsRule.Spec.Enabled = true
	_, err = kube.GetKubeClient().GetDnsClientSet().BinddnsV1().DnsRules().Update(ctx, dnsRule, v1.UpdateOptions{})
	if err != nil {
		zlog.Error(err)
		ctx.JSON(200, &Response{
			Code: ERROR,
			Msg:  err.Error(),
		})
		return
	}
	ctx.JSON(200, &Response{Code: SUCCESS})
}
