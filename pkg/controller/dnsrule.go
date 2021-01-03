package controller

import (
	"context"
	"fmt"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	binddnsv1 "github.com/bind-dns/binddns-operator/pkg/apis/binddns/v1"
	"github.com/bind-dns/binddns-operator/pkg/controller/queue"
	"github.com/bind-dns/binddns-operator/pkg/kube"
	"github.com/bind-dns/binddns-operator/pkg/utils"
	zlog "github.com/bind-dns/binddns-operator/pkg/utils/zaplog"
)

type dnsRuleEventHandler struct {
	dnsController *DnsController
}

func (handler *dnsRuleEventHandler) OnAdd(obj interface{}) {
	rule := obj.(*binddnsv1.DnsRule)

	// Update DnsRule resource status.
	if rule.Status.CreateTime == "" {
		rule.Status = binddnsv1.DnsRuleStatus{
			CreateTime: utils.TimeNow(),
			UpdateTime: utils.TimeNow(),
		}
	}
	if err := updateDnsRuleStatus(rule); err != nil {
		zlog.Errorf("update dnsrules \"%s\" status failed, err: %s", rule.Name, err.Error())
	}

	// Send message to queue.
	if handler.dnsController.enqueue(&queue.DnsEvent{
		Domain:    rule.Spec.Zone,
		EventType: queue.DomainUpdate,
	}) {
		zlog.Infof("dnsrules \"%s\" is added. %s", rule.Name, convertRuleToString(rule))
	}
}

func (handler *dnsRuleEventHandler) OnUpdate(oldObj, newObj interface{}) {
	rule := newObj.(*binddnsv1.DnsRule)

	// Update DnsRule resource status
	rule.Status.UpdateTime = utils.TimeNow()
	if err := updateDnsRuleStatus(rule); err != nil {
		zlog.Errorf("update dnsrules \"%s\" failed, err: %s", rule.Name, err.Error())
	}

	// Send message to queue.
	if handler.dnsController.enqueue(&queue.DnsEvent{
		Domain:    rule.Spec.Zone,
		EventType: queue.DomainUpdate,
	}) {
		zlog.Infof("dnsrules \"%s\" is updated. %s", rule.Name, convertRuleToString(rule))
	}
}

func (handler *dnsRuleEventHandler) OnDelete(obj interface{}) {
	rule := obj.(*binddnsv1.DnsRule)

	// Send message to queue.
	if handler.dnsController.enqueue(&queue.DnsEvent{
		Domain:    rule.Spec.Zone,
		EventType: queue.DomainUpdate,
	}) {
		zlog.Infof("dnsrules \"%s\" is deleted. %s", rule.Name, convertRuleToString(rule))
	}
}

func convertRuleToString(rule *binddnsv1.DnsRule) string {
	return fmt.Sprintf("Zone: %s, Host: %s, Type: %s, Data: %s, Ttl: %d, MxPriority: %d",
		rule.Spec.Zone, rule.Spec.Host, rule.Spec.Type, rule.Spec.Data, rule.Spec.Ttl, rule.Spec.MxPriority)
}

func updateDnsRuleStatus(rule *binddnsv1.DnsRule) (err error) {
	_, err = kube.GetKubeClient().GetDnsClientSet().BinddnsV1().DnsRules(rule.Namespace).
		UpdateStatus(context.Background(), rule, v1.UpdateOptions{})
	return err
}
