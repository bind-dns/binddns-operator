package utils

import "regexp"

var (
	DnsType    = []string{"A", "MX", "CNAME", "NS", "PTR", "TXT", "AAAA", "SRV", "URL"}
	DnsTypeMap = map[string]string{
		"A":     "A",
		"MX":    "MX",
		"CNAME": "CNAME",
		"NS":    "NS",
		"PTR":   "PTR",
		"TXT":   "TXT",
		"AAAA":  "AAAA",
		"SRV":   "SRV",
		"URL":   "URL",
	}
	DomainRegexp   = regexp.MustCompile(`^(([a-zA-Z]{1})|([a-zA-Z]{1}[a-zA-Z]{1})|([a-zA-Z]{1}[0-9]{1})|([0-9]{1}[a-zA-Z]{1})|([a-zA-Z0-9][a-zA-Z0-9-_]{1,61}[a-zA-Z0-9]))\.([a-zA-Z]{2,6}|[a-zA-Z0-9-]{2,30}\.[a-zA-Z]{2,3})$`)
	IPRegexp       = regexp.MustCompile(`^((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?)$`)
	HostnameRegexp = regexp.MustCompile(`^[[:alnum:]][[:alnum:]\-]{0,61}[[:alnum:]]|[[:alpha:]]$`)
	// 详情见互斥关系图， 1 --互斥，0 --不互斥
	DnsRuleRelation = map[string]map[string]int{
		"A":     {"A": 0, "MX": 0, "CNAME": 1, "NS": 0, "TXT": 0, "AAAA": 0, "SRV": 0},
		"MX":    {"A": 0, "MX": 0, "CNAME": 1, "NS": 1, "TXT": 0, "AAAA": 0, "SRV": 0},
		"CNAME": {"A": 1, "MX": 1, "CNAME": 1, "NS": 1, "TXT": 1, "AAAA": 1, "SRV": 0},
		"NS":    {"A": 1, "MX": 1, "CNAME": 1, "NS": 0, "TXT": 1, "AAAA": 1, "SRV": 0},
		"TXT":   {"A": 0, "MX": 0, "CNAME": 1, "NS": 1, "TXT": 0, "AAAA": 0, "SRV": 0},
		"AAAA":  {"A": 0, "MX": 0, "CNAME": 1, "NS": 1, "TXT": 0, "AAAA": 0, "SRV": 0},
		"SRV":   {"A": 0, "MX": 0, "CNAME": 0, "NS": 0, "TXT": 0, "AAAA": 0, "SRV": 0},
	}
)
