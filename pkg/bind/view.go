package bind

import (
	"context"
	"fmt"
	"os"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "github.com/bind-dns/binddns-operator/pkg/apis/binddns/v1"
	"github.com/bind-dns/binddns-operator/pkg/kube"
	"github.com/bind-dns/binddns-operator/pkg/utils"
	zlog "github.com/bind-dns/binddns-operator/pkg/utils/zaplog"
)

// initAllViews used to init the bind9 view config
func (handler *DnsHandler) initAllViews(ctx context.Context) (err error) {
	zlog.Infof("Start to init all views >>>>>>")
	defer func() {
		if err == nil {
			zlog.Infof("Views init successfully.")
		}
	}()

	domains, err := kube.GetKubeClient().GetDnsClientSet().BinddnsV1().DnsDomains().List(ctx, metav1.ListOptions{
		ResourceVersion: "0",
	})
	if err != nil {
		return err
	}

	// init every single view conf file.
	viewList := []string{defaultView}
	allViews := make([]string, 0, len(viewList))
	for i := range viewList {
		if err = handler.initSingleView(viewList[i], domains); err != nil {
			return err
		}
		allViews = append(allViews, fmt.Sprintf("include \"%s\";", handler.ViewDst+"/view_"+defaultView+".conf"))
	}

	file, err := os.Create(handler.ViewDst + "/view.conf")
	defer func() {
		if file != nil {
			_ = file.Close()
		}
	}()
	if err != nil {
		return err
	}
	_, err = file.Write(utils.StringToBytes(strings.Join(allViews, "\n")))
	if err != nil {
		zlog.Error(err)
		return err
	}
	return nil
}

func (handler *DnsHandler) initSingleView(view string, domains *v1.DnsDomainList) error {
	zoneFiles := make([]string, 0, len(domains.Items))
	for i := range domains.Items {
		domain := &domains.Items[i]

		zoneFiles = append(zoneFiles, fmt.Sprintf("        zone %s { type master; file \"%s\"; };",
			domain.Name,
			handler.ZoneDst+"/"+domain.Name+"/db."+view+".conf"),
		)
	}

	zones := strings.Join(zoneFiles, "\n")
	file, err := os.Create(handler.getViewFilePath(view))
	defer func() {
		if file != nil {
			_ = file.Close()
		}
	}()
	if err != nil {
		zlog.Error(err)
		return err
	}

	if view == defaultView {
		_, err = file.Write(utils.StringToBytes(fmt.Sprintf(ViewTemplate, view, "any", zones)))
	} else {
		_, err = file.Write(utils.StringToBytes(fmt.Sprintf(ViewTemplate, view, view, zones)))
	}
	if err != nil {
		zlog.Error(err)
		return err
	}
	return nil
}

func (handler *DnsHandler) getViewFilePath(view string) string {
	return fmt.Sprintf("%s/view_%s.conf", handler.ViewDst, view)
}
