package v1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type DnsDomain struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object metadata.
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the desired behavior of the DnsInstance.
	Spec DnsDomainSpec `json:"spec,omitempty"`

	// Most recently observed status of the DnsInstance.
	Status DnsDomainStatus `json:"status,omitempty"`
}

type DnsDomainSpec struct {
	// Name defines the domain name.
	Name string `json:"name,omitempty"`
	// Enabled defines whether enable the domain.
	// Default true, not required
	Enabled bool `json:"enabled"`
	// Remark defines the remark for the domain, base64 format.
	Remark string `json:"remark"`
}

type DnsDomainStatus struct {
	// CreateTime defines the domain create time.
	CreateTime string `json:"createTime"`
	// UpdateTime defines the domain update time.
	UpdateTime string `json:"updateTime"`
	// Codition defines the
	Condition map[string]DnsDomainCondition `json:"condition"`
}

type DnsDomainCondition struct {
	InstanceName string          `json:"instanceName"`
	Status       ConditionStatus `json:"status"`
}

type ConditionStatus string

const (
	DomainAvailable   ConditionStatus = "Available"
	DomainProgressing ConditionStatus = "Progressing"
	DomainFailure     ConditionStatus = "Failure"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type DnsDomainList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []DnsDomain `json:"items"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type DnsRule struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object metadata.
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the desired behavior of the DnsInstance.
	Spec DnsRuleSpec `json:"spec"`

	// Most recently observed status of the DnsInstance.
	Status DnsRuleStatus `json:"status,omitempty"`
}

type DnsRuleSpec struct {
	Zone       string  `json:"zone,omitempty"`
	Enabled    bool    `json:"enabled"`
	Host       string  `json:"host,omitempty"`
	Type       DnsType `json:"type,omitempty"`
	Data       string  `json:"data,omitempty"`
	Ttl        int32   `json:"ttl,omitempty"`
	MxPriority int32   `json:"maxPriority"`
}

type DnsType string

const (
	TypeA     DnsType = "A"
	TypeMX    DnsType = "MX"
	TypeCNAME DnsType = "CNAME"
	TypeNS    DnsType = "NS"
	TypePTR   DnsType = "PTR"
	TypeTXT   DnsType = "TXT"
)

type DnsRuleStatus struct {
	CreateTime string `json:"createTime"`
	UpdateTime string `json:"updateTime"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type DnsRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []DnsRule `json:"items"`
}
