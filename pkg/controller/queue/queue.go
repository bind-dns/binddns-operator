package queue

import (
	binddnsv1 "github.com/bind-dns/binddns-operator/pkg/apis/binddns/v1"
	"github.com/bind-dns/binddns-operator/pkg/bind"
	k8sstatus "github.com/bind-dns/binddns-operator/pkg/controller/status"
	zlog "github.com/bind-dns/binddns-operator/pkg/utils/zaplog"
)

type DnsEventType string

const (
	DomainAdd    DnsEventType = "DomainAdd"
	DomainUpdate DnsEventType = "DomainUpdate"
	DomainDelete DnsEventType = "DomainDelete"

	defaultQueueSize = 20
)

var (
	domainHandlerMap = map[DnsEventType]func(domain string) error{
		DomainAdd:    bind.GetDnsHandler().ZoneAdd,
		DomainUpdate: bind.GetDnsHandler().ZoneUpdate,
		DomainDelete: bind.GetDnsHandler().ZoneDelete,
	}
)

// DnsEvent used to transfer message from informer watched.
type DnsEvent struct {
	Domain    string       `json:"domain"`
	EventType DnsEventType `json:"eventType"`
}

// DnsQueue defines the queue of informer changed.
type DnsQueue struct {
	name string
	quit chan struct{}
	ch   chan *DnsEvent
}

// NewDnsQueue init an queue object.
func NewDnsQueue(name string) *DnsQueue {
	return &DnsQueue{
		name: name,
		quit: make(chan struct{}),
		ch:   make(chan *DnsEvent, defaultQueueSize),
	}
}

// NewDnsQueueWithSize init an queue object with size.
func NewDnsQueueWithSize(name string, size int32) *DnsQueue {
	return &DnsQueue{
		name: name,
		quit: make(chan struct{}),
		ch:   make(chan *DnsEvent, size),
	}
}

// Enqueue the message of informer changed enqueue.
func (queue *DnsQueue) Enqueue(event *DnsEvent) {
	queue.ch <- event
}

// Run the consumer of queue.
func (queue *DnsQueue) Run() {
	zlog.Infof("Queue [%s] is started.", queue.name)

LOOP:
	for {
		select {
		case event := <-queue.ch:
			{
				zlog.Infof("%s received message: %#v", queue.name, event)
				f, ok := domainHandlerMap[event.EventType]
				if !ok {
					zlog.Errorf("Unknown %s type handler.", event.EventType)
					continue
				}

				err := f(event.Domain)
				if err != nil {
					zlog.Errorf("%s %s failed", event.Domain, event.EventType)
				} else {
					zlog.Infof("%s %s successfully", event.Domain, event.EventType)
				}
				if event.EventType == DomainDelete {
					continue
				}

				status := binddnsv1.DomainAvailable
				if err != nil {
					status = binddnsv1.DomainFailure
				}
				if err = k8sstatus.UpdateDomainStatus(event.Domain, status); err != nil {
					zlog.Error(err)
					continue
				}
				zlog.Infof("%s update status successfully", event.Domain)
			}
		case <-queue.quit:
			{
				break LOOP
			}
		}
	}
	zlog.Infof("Queue [%s] is stopped.", queue.name)
}

func (queue *DnsQueue) Stop() {
	queue.quit <- struct{}{}
}
