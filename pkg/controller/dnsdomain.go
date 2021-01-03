package controller

import (
	"context"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	binddnsv1 "github.com/bind-dns/binddns-operator/pkg/apis/binddns/v1"
	"github.com/bind-dns/binddns-operator/pkg/controller/queue"
	"github.com/bind-dns/binddns-operator/pkg/kube"
	"github.com/bind-dns/binddns-operator/pkg/utils"
	zlog "github.com/bind-dns/binddns-operator/pkg/utils/zaplog"
)

type dnsDomainEventHandler struct {
	dnsController *DnsController
}

func (handler *dnsDomainEventHandler) OnAdd(obj interface{}) {
	domain := obj.(*binddnsv1.DnsDomain)

	// Update DnsDomain resource status.
	if domain.Status.CreateTime == "" {
		domain.Status = binddnsv1.DnsDomainStatus{
			CreateTime: utils.TimeNow(),
			UpdateTime: utils.TimeNow(),
			Condition:  make(map[string]binddnsv1.DnsDomainCondition),
		}

		if err := updateDnsDomainStatus(domain); err != nil {
			zlog.Errorf("update dnsdomains \"%s\" status failed, err: %s", domain.Name, err.Error())
		}
	}

	// Send message to queue.
	if handler.dnsController.enqueue(&queue.DnsEvent{
		Domain:    domain.Name,
		EventType: queue.DomainAdd,
	}) {
		zlog.Infof("dnsdomains \"%s\" is added.", domain.Name)
	}
}

func (handler *dnsDomainEventHandler) OnUpdate(oldObj, newObj interface{}) {
	oldDomain := oldObj.(*binddnsv1.DnsDomain)
	newDomain := newObj.(*binddnsv1.DnsDomain)

	// Update DnsDomain resource status.
	newDomain.Status.UpdateTime = utils.TimeNow()
	if err := updateDnsDomainStatus(newDomain); err != nil {
		zlog.Errorf("update dnsdomains \"%s\" status failed, err: %s", newDomain.Name, err.Error())
	}

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
			zlog.Infof("dnsdomains \"%s\" is updated.", newDomain.Name)
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
		zlog.Infof("dnsdomains \"%s\" is deleted.", domain.Name)
	}
}

func updateDnsDomainStatus(domain *binddnsv1.DnsDomain) (err error) {
	_, err = kube.GetKubeClient().GetDnsClientSet().BinddnsV1().DnsDomains(domain.Namespace).
		UpdateStatus(context.Background(), domain, v1.UpdateOptions{})
	return err
}
