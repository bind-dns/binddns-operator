package webhook

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	v1 "github.com/bind-dns/binddns-operator/pkg/apis/binddns/v1"
	"github.com/bind-dns/binddns-operator/pkg/kube"
	"github.com/bind-dns/binddns-operator/pkg/utils"
	zlog "github.com/bind-dns/binddns-operator/pkg/utils/zaplog"
)

var (
	kindHandlerMap = make(map[string]func(ctx context.Context, request *v1beta1.AdmissionRequest) ([]patchOperation, error))
)

func (server *AdmissionWebhookServer) registerRouter() {
	kindHandlerMap["DnsDomain"] = server.handleDnsDomain
	kindHandlerMap["DnsRule"] = server.handleDnsRule

	server.HttpHandler.Any("/mutate", server.mutate)
}

func (server *AdmissionWebhookServer) webhookAllow(ctx *gin.Context, allowed bool, reqUID types.UID, errMsg string) {
	ctx.JSON(http.StatusOK, &v1beta1.AdmissionReview{
		Response: &v1beta1.AdmissionResponse{
			UID:     reqUID,
			Allowed: allowed,
			Result: &metav1.Status{
				Message: errMsg,
			},
		},
	})
}

func (server *AdmissionWebhookServer) mutate(ctx *gin.Context) {
	ar := new(v1beta1.AdmissionReview)
	if err := ctx.BindJSON(ar); err != nil {
		zlog.Errorf("marshal request body failed, err: ", err.Error())
		server.webhookAllow(ctx, false, "", err.Error())
		return
	}

	// Check request whether is nil.
	req := ar.Request
	if req != nil {
		zlog.Infof("Received request. UID: %s, Operation: %s, Kind: %v.", req.UID, req.Operation, req.Kind)
	} else {
		server.webhookAllow(ctx, false, "", "Unknown Request.")
		return
	}

	// Handle webhook request
	f, ok := kindHandlerMap[req.Kind.Kind]
	if !ok {
		zlog.Errorf("Unknown kind: %s", req.Kind.Kind)
		server.webhookAllow(ctx, false, req.UID, "Unknown Kind: "+req.Kind.Kind)
		return
	}
	ops, err := f(ctx, req)
	if err != nil {
		server.webhookAllow(ctx, false, req.UID, err.Error())
		return
	}

	// Marshal patch data.
	bs, err := json.Marshal(ops)
	if err != nil {
		zlog.Error(err)
		server.webhookAllow(ctx, false, req.UID, err.Error())
		return
	}

	patchType := v1beta1.PatchTypeJSONPatch
	ctx.JSON(http.StatusOK, &v1beta1.AdmissionReview{
		Response: &v1beta1.AdmissionResponse{
			UID:       req.UID,
			Allowed:   true,
			Patch:     bs,
			PatchType: &patchType,
		},
	})
}

func (server *AdmissionWebhookServer) handleDnsDomain(ctx context.Context, req *v1beta1.AdmissionRequest) (ops []patchOperation, err error) {
	if req.Operation != v1beta1.Create && req.Operation != v1beta1.Update {
		return nil, nil
	}
	dnsDomain := new(v1.DnsDomain)
	if err := json.Unmarshal(req.Object.Raw, dnsDomain); err != nil {
		zlog.Errorf("unmarshal DnsDomain kind resource failed, err: %s", err.Error())
		return nil, err
	}
	zlog.Infof("Handle DnsDomain/%s started.", dnsDomain.Name)

	if req.Operation == v1beta1.Create && !utils.DomainRegexp.MatchString(dnsDomain.Name) {
		return nil, errors.Errorf("DnsDomain name format not valid, should be %s", utils.DomainRegexp)
	}
	return nil, nil
}

func (server *AdmissionWebhookServer) handleDnsRule(ctx context.Context, req *v1beta1.AdmissionRequest) (ops []patchOperation, err error) {
	if req.Operation != v1beta1.Create && req.Operation != v1beta1.Update {
		return nil, nil
	}
	dnsRule := new(v1.DnsRule)
	if err = json.Unmarshal(req.Object.Raw, dnsRule); err != nil {
		zlog.Errorf("unmarshal DnsRule kind resource failed, err: %s", err.Error())
		return nil, err
	}
	zlog.Infof("Handle DnsRule/%s started.", dnsRule.Name)

	// If the operation isn't DELETE, confirm that DnsDomain is exist.
	// The DnsRule cannot be changed or added if DnsDomain is deleted.
	_, err = kube.GetKubeClient().GetDnsClientSet().BinddnsV1().DnsDomains().Get(ctx, dnsRule.Spec.Zone,
		metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// Change DnsRule name
	if req.Operation == v1beta1.Create {
		ops = append(ops, updateDnsRuleName(dnsRule))
	}

	// Change DnsRule labels
	ops = append(ops, updateDnsRuleLabels(dnsRule)...)

	// DnsRule common check
	if dataOps, err := checkDnsRuleCommon(dnsRule); err != nil {
		return nil, err
	} else {
		ops = append(ops, dataOps...)
	}

	// Check whether the DnsRule spec.zone is changed.
	isUpdate := req.Operation == v1beta1.Update
	if isUpdate {
		oldRule, err := kube.GetKubeClient().GetDnsClientSet().BinddnsV1().DnsRules().
			Get(ctx, dnsRule.Name, metav1.GetOptions{})
		if err != nil {
			zlog.Error(err)
			return nil, err
		}
		if oldRule.Spec.Zone != dnsRule.Spec.Zone {
			return nil, errors.Errorf("DnsRule spec.zone cannot change when update.")
		}
	}

	// Check exclusion and exist.
	if err := checkDnsTypeExclusion(ctx, dnsRule, isUpdate); err != nil {
		return nil, err
	}

	return ops, nil
}

func checkDnsRuleCommon(rule *v1.DnsRule) (ops []patchOperation, err error) {
	host := strings.TrimSpace(rule.Spec.Host)
	if host == "" {
		return nil, errors.Errorf("DnsRule spec.host cannot be empty.")
	}
	if !(host == "*" || host == "@") && !utils.HostnameRegexp.MatchString(host) {
		return nil, errors.Errorf("DnsRule spec.host format not valid, should be %s.", utils.HostnameRegexp)
	}

	dnsType := strings.TrimSpace(string(rule.Spec.Type))
	if dnsType == "" {
		return nil, errors.Errorf("DnsRule spec.type cannot be empty.")
	}
	if _, ok := utils.DnsTypeMap[dnsType]; !ok {
		return nil, errors.Errorf("DnsRule spec.type(%s) must be in %s.", dnsType, utils.DnsType)
	}

	if rule.Spec.Ttl < 1 {
		return nil, errors.Errorf("DnsRule spec.ttl must > 1.")
	}

	if dnsType == "MX" && (rule.Spec.MxPriority <= 0 || rule.Spec.MxPriority >= 100) {
		return nil, errors.Errorf("DnsRule spec.type should in 1-100.")
	}

	data := strings.TrimSpace(rule.Spec.Data)
	if dnsType == "A" {
		if !utils.IPRegexp.MatchString(data) {
			return nil, errors.Errorf("DnsRule spec.data with A type should be %s", utils.IPRegexp)
		}
	} else if dnsType == "CNAME" || dnsType == "NS" || dnsType == "MX" {
		length := len(data)

		// FIXME Ignore check data whether is a normal domain.
		if data[length-1] != '.' {
			data = data + "."
			ops = []patchOperation{
				{
					Op:    "replace",
					Path:  "/spec/data",
					Value: data,
				},
			}
		}
	}
	return ops, nil
}

func checkDnsTypeExclusion(ctx context.Context, rule *v1.DnsRule, isUpdate bool) error {
	relation, ok := utils.DnsRuleRelation[string(rule.Spec.Type)]
	if !ok {
		return nil
	}
	exclusionTypes := make([]string, 0, len(relation))
	for k, v := range relation {
		if v == 1 {
			exclusionTypes = append(exclusionTypes, k)
		}
	}

	rules, err := kube.GetKubeClient().GetDnsClientSet().BinddnsV1().DnsRules().List(ctx, metav1.ListOptions{
		ResourceVersion: "0",
		LabelSelector: strings.Join([]string{
			utils.LabelZoneDnsRule + "=" + rule.Spec.Zone,
			utils.LabelHostDnsRule + "=" + rule.Spec.Host,
		}, ","),
	})
	if err != nil {
		zlog.Error(err)
		return err
	}
	for i := range rules.Items {
		item := &rules.Items[i]

		// Check exclusion
		if isUpdate {
			if item.Name != rule.Name && utils.SliceContain(exclusionTypes, string(item.Spec.Type)) {
				return errors.Errorf("Exclusion with %s(Host: %s, Type: %s, Data: %s)",
					item.Name, item.Spec.Host, item.Spec.Type, item.Spec.Data)
			}
		} else {
			if utils.SliceContain(exclusionTypes, string(item.Spec.Type)) {
				return errors.Errorf("Exclusion with %s(Host: %s, Type: %s, Data: %s)",
					item.Name, item.Spec.Host, item.Spec.Type, item.Spec.Data)
			}
		}

		// Check existed
		if item.Spec.Type == rule.Spec.Type && item.Spec.Data == rule.Spec.Data {
			if !isUpdate || (isUpdate && item.Name != rule.Name) {
				return errors.Errorf("Same with %s(Host: %s, Type: %s, Data: %s)",
					item.Name, item.Spec.Host, item.Spec.Type, item.Spec.Data)
			}
		}
	}
	return nil
}
