package webhook

import (
	"strings"

	uuid "github.com/satori/go.uuid"

	binddnsv1 "github.com/bind-dns/binddns-operator/pkg/apis/binddns/v1"
	"github.com/bind-dns/binddns-operator/pkg/utils"
)

type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

func updateDnsRuleName(rule *binddnsv1.DnsRule) patchOperation {
	return patchOperation{
		Op:   "replace",
		Path: "/metadata/name",
		Value: rule.Spec.Zone + "-" +
			utils.SubString(strings.ReplaceAll(uuid.NewV4().String(), "-", ""), 0, 8),
	}
}

func updateDnsRuleLabels(rule *binddnsv1.DnsRule) (ops []patchOperation) {
	labels := map[string]string{
		utils.LabelHostDnsRule: rule.Spec.Host,
		utils.LabelZoneDnsRule: rule.Spec.Zone,
		utils.LabelTypeDnsRule: string(rule.Spec.Type),
	}

	if rule.ObjectMeta.Labels == nil || len(rule.ObjectMeta.Labels) == 0 {
		ops = append(ops, patchOperation{
			Op:    "add",
			Path:  "/metadata/labels",
			Value: labels,
		})
	} else {
		for k, v := range labels {
			ops = append(ops, patchOperation{
				Op:    "replace",
				Path:  "/metadata/labels/" + escapeSlash(k),
				Value: v,
			})
		}
	}
	return ops
}

func escapeSlash(k string) string {
	return strings.ReplaceAll(k, "/", "~1")
}
