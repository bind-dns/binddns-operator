package bind

const (
	// ZoneTemplate defines the zone template.
	ZoneTemplate = `
$TTL    60
$ORIGIN %s.
@       86400   SOA   ns1.%s. admin.%s.(
                      %d                ; Serial
                      3600              ; Refresh period
                      3600              ; Retry period
                      86400             ; Expire period
                      3600 )            ; Minimum TTL

@       86400     NS     ns1.%s.
@       86400     NS     ns2.%s.
@       86400     NS     ns3.%s.

; Hosts
%s
`

	// ViewTemplate defines the view template
	ViewTemplate = `
view "%s" {
        match-clients {%s; %s; };
        allow-query-cache       { none; };
        allow-transfer          { none; };
        allow-recursion         { none; };
%s
};
`

	// defaultDnsConfDir
	defaultDnsConfDir = "/etc/named"
)

var (
	// Global singleton object.
	globalHandler *DnsHandler
)

type DnsHandler struct {
	ZoneDst string `json:"zoneDst"`
	ViewDst string `json:"viewDst"`
	AclDst  string `json:"aclDst"`
}

func NewDnsHandler() *DnsHandler {
	globalHandler = &DnsHandler{
		ZoneDst: defaultDnsConfDir + "/zone",
		ViewDst: defaultDnsConfDir + "/view",
		AclDst:  defaultDnsConfDir + "/acl",
	}
	return globalHandler
}

func GetDnsHandler() *DnsHandler {
	if globalHandler == nil {
		return NewDnsHandler()
	}
	return globalHandler
}
