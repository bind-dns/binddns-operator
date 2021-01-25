package router

import (
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	k8sErrors "k8s.io/apimachinery/pkg/api/errors"

	binddnsv1 "github.com/bind-dns/binddns-operator/pkg/apis/binddns/v1"
	"github.com/bind-dns/binddns-operator/pkg/kube"
	zlog "github.com/bind-dns/binddns-operator/pkg/utils/zaplog"
)

type DnsDomainRequest struct {
	Name   string `json:"name"`
	Remark string `json:"remark,omitempty"`
}

type DnsDomainEntity struct {
	Name       string
	CreateTime string
	Status     string
	Enabled    bool
	Remark     string
}

func createDomain(ctx *gin.Context) {
	req := new(DnsDomainRequest)
	if err := ctx.BindJSON(req); err != nil {
		zlog.Error(err)
		ctx.JSON(http.StatusOK, &Response{
			Code: ERROR,
			Msg:  err.Error(),
		})
		return
	}
	_, err := kube.GetKubeClient().GetDnsClientSet().BinddnsV1().DnsDomains().Create(ctx, &binddnsv1.DnsDomain{
		TypeMeta: v1.TypeMeta{
			Kind:       "DnsDomain",
			APIVersion: "binddns.github.com/v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: req.Name,
		},
		Spec: binddnsv1.DnsDomainSpec{
			Enabled: true,
			Remark:  req.Remark,
		},
	}, v1.CreateOptions{})
	if err != nil {
		ctx.JSON(200, &Response{
			Code: ERROR,
			Msg:  err.Error(),
		})
		return
	}
	ctx.JSON(200, &Response{
		Code: SUCCESS,
	})
}

func deleteDomain(ctx *gin.Context) {
	domain := ctx.Param("domain")
	err := kube.GetKubeClient().GetDnsClientSet().BinddnsV1().DnsDomains().Delete(ctx, domain, v1.DeleteOptions{})
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

func pauseDomain(ctx *gin.Context) {
	domain := ctx.Param("domain")

	dnsDomain, err := kube.GetKubeClient().GetDnsClientSet().BinddnsV1().DnsDomains().Get(ctx, domain, v1.GetOptions{})
	if err != nil {
		zlog.Error(err)
		ctx.JSON(200, &Response{
			Code: ERROR,
			Msg:  err.Error(),
		})
		return
	}
	if !dnsDomain.Spec.Enabled {
		ctx.JSON(200, &Response{Code: SUCCESS})
		return
	}

	dnsDomain.Spec.Enabled = false
	_, err = kube.GetKubeClient().GetDnsClientSet().BinddnsV1().DnsDomains().Update(ctx, dnsDomain, v1.UpdateOptions{})
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

func openDomain(ctx *gin.Context) {
	domain := ctx.Param("domain")

	dnsDomain, err := kube.GetKubeClient().GetDnsClientSet().BinddnsV1().DnsDomains().Get(ctx, domain, v1.GetOptions{})
	if err != nil {
		zlog.Error(err)
		ctx.JSON(200, &Response{
			Code: ERROR,
			Msg:  err.Error(),
		})
		return
	}
	if dnsDomain.Spec.Enabled {
		ctx.JSON(200, &Response{Code: SUCCESS})
		return
	}

	dnsDomain.Spec.Enabled = true
	_, err = kube.GetKubeClient().GetDnsClientSet().BinddnsV1().DnsDomains().Update(ctx, dnsDomain, v1.UpdateOptions{})
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

func listDomains(ctx *gin.Context) {
	search := strings.TrimSpace(ctx.Query("search"))
	list, err := kube.GetKubeClient().GetDnsClientSet().BinddnsV1().DnsDomains().List(ctx, v1.ListOptions{
		ResourceVersion: "0",
	})
	if err != nil {
		zlog.Error(err)
		ctx.JSON(200, &Response{
			Code: ERROR,
			Msg:  err.Error(),
		})
		return
	}

	s := make([]*DnsDomainEntity, 0, len(list.Items))
	for i := range list.Items {
		item := &list.Items[i]

		if search == "" || strings.Contains(item.Name, search) {
			s = append(s, &DnsDomainEntity{
				Name:       item.Name,
				CreateTime: item.Status.CreateTime,
				Status:     "Available",
				Enabled:    item.Spec.Enabled,
				Remark:     item.Spec.Remark,
			})
		}
	}

	// Sort by createTime
	sort.Sort(DnsDomainSort(s))

	ctx.JSON(200, &Response{
		Code: SUCCESS,
		Data: s,
	})
}
