package v1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +k8s:openapi-gen=true
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="enabled",type=boolean,JSONPath=`.spec.enabled`
// +kubebuilder:printcolumn:name="remark",type=string,JSONPath=`.spec.remark`
// +kubebuilder:printcolumn:name="update",type=string,JSONPath=`.status.updateTime`
// +kubebuilder:printcolumn:name="status",type=string,JSONPath=`.status.phase`
type DnsDomain struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object metadata.
	metav1.ObjectMeta `json:"metadata"`

	// Specification of the desired behavior of the DnsDomain.
	Spec DnsDomainSpec `json:"spec"`

	// +kubebuilder:validation:Optional
	// Most recently observed status of the DnsDomain.
	Status DnsDomainStatus `json:"status,omitempty"`
}

type DnsDomainSpec struct {
	// Enabled defines whether enable the domain.
	Enabled bool `json:"enabled"`

	// +kubebuilder:validation:Optional
	// Remark defines the remark for the domain, base64 format.
	Remark string `json:"remark,omitempty"`
}

type DnsDomainStatus struct {
	// +kubebuilder:validation:Optional
	// CreateTime defines the domain create time.
	CreateTime string `json:"createTime"`

	// +kubebuilder:validation:Optional
	// UpdateTime defines the domain update time.
	UpdateTime string `json:"updateTime"`

	// +kubebuilder:validation:Optional
	// InstanceStatuses defines the domain status of every instance
	InstanceStatuses map[string]InstanceStatus `json:"instanceStatuses"`

	// +kubebuilder:validation:Optional
	Phase DomainStatus `json:"phase"`
}

type InstanceStatus struct {
	// +kubebuilder:validation:Optional
	Name string `json:"name"`

	// +kubebuilder:validation:Optional
	Status DomainStatus `json:"status"`

	// +kubebuilder:validation:Optional
	UpdatedAt string `json:"updatedAt"`
}

type DomainStatus string

const (
	DomainAvailable   DomainStatus = "Available"
	DomainProgressing DomainStatus = "Progressing"
	DomainFailure     DomainStatus = "Failure"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type DnsDomainList struct {
	metav1.TypeMeta `json:",inline"`

	metav1.ListMeta `json:"metadata"`

	Items []DnsDomain `json:"items"`
}

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// +k8s:openapi-gen=true
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Zone",type=string,JSONPath=`.spec.zone`
// +kubebuilder:printcolumn:name="Enabled",type=boolean,JSONPath=`.spec.enabled`
// +kubebuilder:printcolumn:name="Host",type=string,JSONPath=`.spec.host`
// +kubebuilder:printcolumn:name="Type",type=string,JSONPath=`.spec.type`
// +kubebuilder:printcolumn:name="Data",type=string,JSONPath=`.spec.data`
// +kubebuilder:printcolumn:name="Ttl",type=number,JSONPath=`.spec.ttl`
// +kubebuilder:printcolumn:name="MxPriority",type=number,JSONPath=`.spec.mxPriority`
type DnsRule struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object metadata.
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Specification of the desired behavior of the DnsRule.
	Spec DnsRuleSpec `json:"spec"`

	// +kubebuilder:validation:Optional
	// Most recently observed status of the DnsRule.
	Status DnsRuleStatus `json:"status,omitempty"`
}

type DnsRuleSpec struct {
	Zone    string  `json:"zone"`
	Enabled bool    `json:"enabled"`
	Host    string  `json:"host"`
	Type    DnsType `json:"type"`
	Data    string  `json:"data"`
	Ttl     int32   `json:"ttl"`

	// +kubebuilder:validation:Optional
	MxPriority int32 `json:"mxPriority"`
}

type DnsType string

const (
	TypeA     DnsType = "A"
	TypeMX    DnsType = "MX"
	TypeCNAME DnsType = "CNAME"
	TypeNS    DnsType = "NS"
	TypePTR   DnsType = "PTR"
	TypeTXT   DnsType = "TXT"
	TypeAAAA  DnsType = "AAAA"
	TypeSRV   DnsType = "SRV"
	TypeURL   DnsType = "URL"
)

type DnsRuleStatus struct {
	// +kubebuilder:validation:Optional
	CreateTime string `json:"createTime"`
	// +kubebuilder:validation:Optional
	UpdateTime string `json:"updateTime"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type DnsRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []DnsRule `json:"items"`
}
