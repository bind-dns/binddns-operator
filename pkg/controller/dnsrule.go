package controller

import (
	"context"
	"fmt"

	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	binddnsv1 "github.com/bind-dns/binddns-operator/pkg/apis/binddns/v1"
	"github.com/bind-dns/binddns-operator/pkg/controller/queue"
	k8sstatus "github.com/bind-dns/binddns-operator/pkg/controller/status"
	"github.com/bind-dns/binddns-operator/pkg/kube"
	zlog "github.com/bind-dns/binddns-operator/pkg/utils/zaplog"
)

type dnsRuleEventHandler struct {
	dnsController *DnsController
}

func (handler *dnsRuleEventHandler) OnAdd(obj interface{}) {
	rule := obj.(*binddnsv1.DnsRule)

	// Send message to queue.
	if handler.dnsController.enqueue(&queue.DnsEvent{
		Domain:    rule.Spec.Zone,
		EventType: queue.DomainUpdate,
	}) {
		// Update DnsRule status
		if err := k8sstatus.UpdateRuleStatus(rule.Name); err != nil {
			zlog.Error(err)
		}

		zlog.Infof("DnsRule \"%s\" is added. %s", rule.Name, convertRuleToString(rule))
	}
}

func (handler *dnsRuleEventHandler) OnUpdate(oldObj, newObj interface{}) {
	oldRule := oldObj.(*binddnsv1.DnsRule)
	newRule := newObj.(*binddnsv1.DnsRule)

	isUpdate := false
	if oldRule.Spec.Enabled != newRule.Spec.Enabled || oldRule.Spec.Data != newRule.Spec.Data ||
		oldRule.Spec.Ttl != newRule.Spec.Ttl || oldRule.Spec.Host != newRule.Spec.Host ||
		oldRule.Spec.Type != newRule.Spec.Type || oldRule.Spec.MxPriority != newRule.Spec.MxPriority {
		isUpdate = true
	}
	if !isUpdate {
		return
	}

	// Send message to queue.
	if handler.dnsController.enqueue(&queue.DnsEvent{
		Domain:    newRule.Spec.Zone,
		EventType: queue.DomainUpdate,
	}) {
		// Update DnsRule status
		if err := k8sstatus.UpdateRuleStatus(newRule.Name); err != nil {
			zlog.Error(err)
		}

		zlog.Infof("DnsRule \"%s\" is updated. %s", newRule.Name, convertRuleToString(newRule))
	}
}

func (handler *dnsRuleEventHandler) OnDelete(obj interface{}) {
	rule := obj.(*binddnsv1.DnsRule)

	zone := rule.Spec.Zone
	// If DnsDomain is deleted, don't handle the event of DnsRule delete.
	_, err := kube.GetKubeClient().GetDnsClientSet().BinddnsV1().DnsDomains().Get(context.Background(), zone, v1.GetOptions{})
	if err != nil {
		if k8sErrors.IsNotFound(err) {
			zlog.Warnf("DnsDomain \"%s\" is not found", zone)
		} else {
			zlog.Error(err)
		}
		return
	}

	// Send message to queue.
	if handler.dnsController.enqueue(&queue.DnsEvent{
		Domain:    zone,
		EventType: queue.DomainUpdate,
	}) {
		// Update DnsRule status
		if err := k8sstatus.UpdateDomainStatus(zone, binddnsv1.DomainProgressing); err != nil {
			zlog.Error(err)
		}

		zlog.Infof("DnsRule \"%s\" is deleted. %s", rule.Name, convertRuleToString(rule))
	}
}

func convertRuleToString(rule *binddnsv1.DnsRule) string {
	return fmt.Sprintf("Zone: %s, Host: %s, Type: %s, Data: %s, Ttl: %d, MxPriority: %d",
		rule.Spec.Zone, rule.Spec.Host, rule.Spec.Type, rule.Spec.Data, rule.Spec.Ttl, rule.Spec.MxPriority)
}
