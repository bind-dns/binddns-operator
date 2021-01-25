package bind

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	binddnsv1 "github.com/bind-dns/binddns-operator/pkg/apis/binddns/v1"
	"github.com/bind-dns/binddns-operator/pkg/kube"
	"github.com/bind-dns/binddns-operator/pkg/utils"
	zlog "github.com/bind-dns/binddns-operator/pkg/utils/zaplog"
)

// initAllZones used to init the bind9 zone config
func (handler *DnsHandler) initAllZones(ctx context.Context) (err error) {
	zlog.Infof("Start to init all zones >>>>>>")
	defer func() {
		if err == nil {
			zlog.Infof("Zones init successfully.")
		}
	}()

	domains, err := kube.GetKubeClient().GetDnsClientSet().BinddnsV1().DnsDomains().List(ctx, v1.ListOptions{
		ResourceVersion: "0",
	})
	if err != nil {
		zlog.Error(err)
		return err
	}
	for i := range domains.Items {
		err = handler.initZone(ctx, &domains.Items[i])
		if err != nil {
			return err
		}
	}
	return nil
}

// ZoneAdd used to add a zone.
func (handler *DnsHandler) ZoneAdd(zone string) error {
	ctx := context.Background()
	domain, err := kube.GetKubeClient().GetDnsClientSet().BinddnsV1().DnsDomains().Get(ctx, zone, v1.GetOptions{})
	if err != nil {
		zlog.Error(err)
		return err
	}
	if err := handler.initZone(ctx, domain); err != nil {
		return err
	}

	// There is only one default view.
	if err := exec.Command("/etc/named/rndc", "addzone", zone, "IN", defaultView,
		fmt.Sprintf("{ type master; file \"%s\";};", handler.getZoneFilePath(zone, defaultView))).Run(); err != nil {
		zlog.Error(err)
		return err
	}
	return nil
}

// ZoneUpdate used to update a zone.
func (handler *DnsHandler) ZoneUpdate(zone string) error {
	ctx := context.Background()
	domain, err := kube.GetKubeClient().GetDnsClientSet().BinddnsV1().DnsDomains().Get(ctx, zone, v1.GetOptions{})
	if err != nil {
		zlog.Error(err)
		return err
	}
	if err := handler.initZone(ctx, domain); err != nil {
		return err
	}

	// There is only one default view.
	cmd := exec.Command("/etc/named/rndc", "freeze", zone, "IN", defaultView)
	if err := cmd.Run(); err != nil {
		zlog.Error(cmd)
		return err
	}
	cmd = exec.Command("/etc/named/rndc", "reload", zone, "IN", defaultView)
	if err := cmd.Run(); err != nil {
		zlog.Error(cmd)
		return err
	}
	cmd = exec.Command("/etc/named/rndc", "thaw", zone, "IN", defaultView)
	if err := cmd.Run(); err != nil {
		zlog.Error(cmd)
		return err
	}
	return nil
}

// ZoneDelete used to delete a zone.
func (handler *DnsHandler) ZoneDelete(zone string) error {
	if err := os.RemoveAll(handler.getZoneDir(zone)); err != nil {
		zlog.Error(err)
		return err
	}

	views := []string{defaultView}
	for _, view := range views {
		if err := exec.Command("/etc/named/rndc", "delzone", zone, "IN", view).Run(); err != nil {
			zlog.Error(err)
			return err
		}
	}
	return nil
}

// initZone will init a single zone config file.
func (handler *DnsHandler) initZone(ctx context.Context, domain *binddnsv1.DnsDomain) error {
	var records []string
	if domain.Spec.Enabled {
		rules, err := kube.GetKubeClient().GetDnsClientSet().BinddnsV1().DnsRules().List(ctx, v1.ListOptions{
			ResourceVersion: "0",
			LabelSelector:   utils.LabelZoneDnsRule + "=" + domain.Name,
		})
		if err != nil {
			zlog.Error(err)
			return err
		}

		for i := range rules.Items {
			item := &rules.Items[i]
			if !item.Spec.Enabled {
				continue
			}
			if item.Spec.Type == "MX" {
				records = append(records, fmt.Sprintf("%s %d %s 10 %s \n", strings.TrimSpace(item.Spec.Host),
					item.Spec.Ttl, item.Spec.Type, item.Spec.Data))
				continue
			}
			records = append(records, fmt.Sprintf("%s %d %s %s\n", strings.TrimSpace(item.Spec.Host),
				item.Spec.Ttl, item.Spec.Type, item.Spec.Data))
		}
	}

	if err := os.MkdirAll(handler.getZoneDir(domain.Name), 0777); err != nil {
		return err
	}
	// There is only one default view.
	file, err := os.Create(handler.getZoneFilePath(domain.Name, defaultView))
	defer func() {
		if file != nil {
			file.Close()
		}
	}()
	if err != nil {
		zlog.Error(err)
		return err
	}

	_, err = file.Write(utils.StringToBytes(fmt.Sprintf(ZoneTemplate, domain.Name, time.Now().Unix(), strings.Join(records, "\n"))))
	if err != nil {
		return err
	}
	return nil
}

func (handler *DnsHandler) getZoneDir(zone string) string {
	return fmt.Sprintf("%s/%s", handler.ZoneDst, zone)
}

func (handler *DnsHandler) getZoneFilePath(zone, view string) string {
	return fmt.Sprintf("%s/%s/db.%s.conf", handler.ZoneDst, zone, view)
}
