package k8sstatus

import (
	"context"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	binddnsv1 "github.com/bind-dns/binddns-operator/pkg/apis/binddns/v1"
	"github.com/bind-dns/binddns-operator/pkg/kube"
	"github.com/bind-dns/binddns-operator/pkg/utils"
)

func UpdateDomainStatus(zone string, status binddnsv1.DomainStatus) error {
	domain, err := kube.GetKubeClient().GetDnsClientSet().BinddnsV1().DnsDomains().Get(context.Background(), zone, v1.GetOptions{})
	if err != nil {
		return err
	}

	if domain.Status.CreateTime == "" {
		domain.Status.CreateTime = utils.TimeNow()
	}
	domain.Status.UpdateTime = utils.TimeNow()
	if domain.Status.InstanceStatuses == nil {
		domain.Status.InstanceStatuses = make(map[string]binddnsv1.InstanceStatus)
	}

	podName := utils.GetPodName()
	domain.Status.Phase = status
	domain.Status.InstanceStatuses[podName] = binddnsv1.InstanceStatus{
		Status: status,
		Name: podName,
		UpdatedAt: utils.TimeNow(),
	}

	_, err = kube.GetKubeClient().GetDnsClientSet().BinddnsV1().DnsDomains().
		UpdateStatus(context.Background(), domain, v1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func UpdateRuleStatus(name string) error {
	rule, err := kube.GetKubeClient().GetDnsClientSet().BinddnsV1().DnsRules().Get(context.Background(), name, v1.GetOptions{})
	if err != nil {
		return err
	}

	if rule.Status.CreateTime == "" {
		rule.Status.CreateTime = utils.TimeNow()
	}
	rule.Status.UpdateTime = utils.TimeNow()
	_, err = kube.GetKubeClient().GetDnsClientSet().BinddnsV1().DnsRules().
		UpdateStatus(context.Background(), rule, v1.UpdateOptions{})
	if err != nil {
		return err
	}

	return UpdateDomainStatus(rule.Spec.Zone, binddnsv1.DomainProgressing)
}
