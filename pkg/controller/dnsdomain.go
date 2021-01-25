package controller

import (
	"context"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	binddnsv1 "github.com/bind-dns/binddns-operator/pkg/apis/binddns/v1"
	"github.com/bind-dns/binddns-operator/pkg/controller/queue"
	"github.com/bind-dns/binddns-operator/pkg/controller/status"
	"github.com/bind-dns/binddns-operator/pkg/kube"
	"github.com/bind-dns/binddns-operator/pkg/utils"
	zlog "github.com/bind-dns/binddns-operator/pkg/utils/zaplog"
)

type dnsDomainEventHandler struct {
	dnsController *DnsController
}

func (handler *dnsDomainEventHandler) OnAdd(obj interface{}) {
	domain := obj.(*binddnsv1.DnsDomain)

	// Send message to queue.
	if handler.dnsController.enqueue(&queue.DnsEvent{
		Domain:    domain.Name,
		EventType: queue.DomainAdd,
	}) {
		// Update DnsDomain status
		if err := k8sstatus.UpdateDomainStatus(domain.Name, binddnsv1.DomainProgressing); err != nil {
			zlog.Error(err)
		}

		zlog.Infof("DnsDomain \"%s\" is added.", domain.Name)
	}
}

func (handler *dnsDomainEventHandler) OnUpdate(oldObj, newObj interface{}) {
	oldDomain := oldObj.(*binddnsv1.DnsDomain)
	newDomain := newObj.(*binddnsv1.DnsDomain)

	// Send DomainEnable message to queue.
	if oldDomain.Spec.Enabled != newDomain.Spec.Enabled {
		if newDomain.Spec.Enabled {
			zlog.Infof("dnsdomains \"%s\" is enabled.", newDomain.Name)
		} else {
			zlog.Infof("dnsdomains \"%s\" is disabled.", newDomain.Name)
		}

		// Send message to queue.
		if handler.dnsController.enqueue(&queue.DnsEvent{
			Domain:    newDomain.Name,
			EventType: queue.DomainUpdate,
		}) {
			// Update DnsDomain status
			if err := k8sstatus.UpdateDomainStatus(newDomain.Name, binddnsv1.DomainProgressing); err != nil {
				zlog.Error(err)
			}

			zlog.Infof("DnsDomain \"%s\" is updated.", newDomain.Name)
		}
	}
}

func (handler *dnsDomainEventHandler) OnDelete(obj interface{}) {
	domain := obj.(*binddnsv1.DnsDomain)

	// Send message to queue.
	if handler.dnsController.enqueue(&queue.DnsEvent{
		Domain:    domain.Name,
		EventType: queue.DomainDelete,
	}) {
		zlog.Infof("DnsDomain \"%s\" is deleted.", domain.Name)

		// Cascade delete DnsRules
		err := kube.GetKubeClient().GetDnsClientSet().BinddnsV1().DnsRules().DeleteCollection(
			context.Background(),
			v1.DeleteOptions{}, v1.ListOptions{
				ResourceVersion: "0",
				LabelSelector:   utils.LabelZoneDnsRule + "=" + domain.Name,
			})
		if err != nil {
			zlog.Error(err)
		}
	}
}
