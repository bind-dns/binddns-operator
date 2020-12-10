# binddns-operator
BindDns Operator creates/configures/manages bind9 dns atop Kubernetes

## 架构

定义两个类型的 CRD: DnsInstance 和 DnsRules。

DnsInstance 使用 Operator 控制 Controller 实例数目
DnsRules 使用 Controller 控制具体的 DNS 规则。

