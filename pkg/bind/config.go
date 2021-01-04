package bind

import (
	"context"
)

// InitConfig
func (handler *DnsHandler) InitConfig(ctx context.Context) (err error) {
	if err = handler.initAllZones(ctx); err != nil {
		return err
	}
	if err = handler.initAllViews(ctx); err != nil {
		return err
	}
	return nil
}

// initAcl used to init the bind9 acl config
func (handler *DnsHandler) initAcl() {
	// doNothing
}
