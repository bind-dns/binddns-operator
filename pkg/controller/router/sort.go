package router

type DnsDomainSort []*DnsDomainEntity

func (ds DnsDomainSort) Len() int {
	return len(ds)
}

func (ds DnsDomainSort) Swap(i, j int) {
	ds[i], ds[j] = ds[j], ds[i]
}

func (ds DnsDomainSort) Less(i, j int) bool {
	return ds[i].CreateTime < ds[j].CreateTime
}

type DnsRuleSort []*DnsRuleEntity

func (rs DnsRuleSort) Len() int {
	return len(rs)
}

func (rs DnsRuleSort) Swap(i, j int) {
	rs[i], rs[j] = rs[j], rs[i]
}

func (rs DnsRuleSort) Less(i, j int) bool {
	return rs[i].CreateTime < rs[j].CreateTime
}
