package bind

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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

// ZoneAdd used to add a zone.
func (handler *DnsHandler) ZoneAdd(zone string) error {
	if err := handler.initZone(context.Background(), zone); err != nil {
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
	if err := handler.initZone(context.Background(), zone); err != nil {
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
	return nil
}

// initZone will init a single zone config file.
func (handler *DnsHandler) initZone(ctx context.Context, zone string) error {
	rules, err := kube.GetKubeClient().GetDnsClientSet().BinddnsV1().DnsRules("").List(ctx, v1.ListOptions{
		ResourceVersion: "0",
	})
	if err != nil {
		zlog.Error(err)
		return err
	}

	var recordStr string
	for i := range rules.Items {
		item := &rules.Items[i]
		if item.Spec.Type == "MX" {
			recordStr = fmt.Sprintf("%s %d %s 10 %s \n", strings.TrimSpace(item.Spec.Host),
				item.Spec.Ttl, item.Spec.Type, item.Spec.Data)
			continue
		}
		recordStr = fmt.Sprintf("%s %d %s %s\n", strings.TrimSpace(item.Spec.Host),
			item.Spec.Ttl, item.Spec.Type, item.Spec.Data)
	}

	if err := os.MkdirAll(handler.getZoneDir(zone), 0777); err != nil {
		return err
	}
	// There is only one default view.
	file, err := os.Create(handler.getZoneFilePath(zone, defaultView))
	defer func() {
		if file != nil {
			file.Close()
		}
	}()
	if err != nil {
		zlog.Error(err)
		return err
	}

	_, err = file.Write(utils.StringToBytes(fmt.Sprintf(ZoneTemplate, zone, time.Now().Unix(), recordStr)))
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
