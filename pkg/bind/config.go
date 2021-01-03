package bind

import (
	"context"
	"fmt"
	"os"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/bind-dns/binddns-operator/pkg/kube"
	"github.com/bind-dns/binddns-operator/pkg/utils"
	zlog "github.com/bind-dns/binddns-operator/pkg/utils/zaplog"
)

// InitConfig
func (handler *DnsHandler) InitConfig(ctx context.Context) (err error) {
	if err = handler.allZones(ctx); err != nil {
		return err
	}
	if err = handler.allViews(ctx); err != nil {
		return err
	}
	return nil
}

// initAcl used to init the bind9 acl config
func (handler *DnsHandler) initAcl() {
	// doNothing
}

// allZones used to init the bind9 zone config
func (handler *DnsHandler) allZones(ctx context.Context) (err error) {
	zlog.Infof("Start to init all zones >>>>>>")
	defer func() {
		if err == nil {
			zlog.Infof("Zones init successfully.")
		}
	}()

	domains, err := kube.GetKubeClient().GetDnsClientSet().BinddnsV1().DnsDomains("").List(ctx, v1.ListOptions{
		ResourceVersion: "0",
	})
	if err != nil {
		zlog.Error(err)
		return err
	}
	for i := range domains.Items {
		err = handler.initZone(ctx, domains.Items[i].Name)
		if err != nil {
			return err
		}
	}
	return nil
}

// allViews used to init the bind9 view config
func (handler *DnsHandler) allViews(ctx context.Context) (err error) {
	zlog.Infof("Start to init all views >>>>>>")
	defer func() {
		if err == nil {
			zlog.Infof("Views init successfully.")
		}
	}()

	// There is only one default acl. After dns view feature add, there will
	// be completed.
	if err = handler.ViewAdd(ctx, "default"); err != nil {
		return err
	}

	allViews := fmt.Sprintf("include \"%s\";\n", handler.ViewDst + "/view_default.conf")

	file, err := os.Create(handler.ViewDst + "/view.conf")
	defer func() {
		if file != nil {
			file.Close()
		}
	}()
	if err != nil {
		return err
	}
	_, err = file.Write(utils.StringToBytes(allViews))
	if err != nil {
		return err
	}
	return nil
}
