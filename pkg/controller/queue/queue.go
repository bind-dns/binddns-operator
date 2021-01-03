package queue

import (
	"github.com/bind-dns/binddns-operator/pkg/bind"
	zlog "github.com/bind-dns/binddns-operator/pkg/utils/zaplog"
)

type DnsEventType string

const (
	DomainAdd    DnsEventType = "DomainAdd"
	DomainUpdate DnsEventType = "DomainUpdate"
	DomainDelete DnsEventType = "DomainDelete"

	defaultQueueSize = 10
)

var (
	domainHandlerMap = map[DnsEventType]func(domain string){
		DomainAdd:    handleDomainAdd,
		DomainUpdate: handleDomainUpdate,
		DomainDelete: handleDomainDelete,
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
				f(event.Domain)
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

func handleDomainAdd(domain string) {
	if err := bind.GetDnsHandler().ZoneAdd(domain); err != nil {
		zlog.Errorf("add domain %s failed, err: %s", domain, err.Error())
	}
}

func handleDomainUpdate(domain string) {
	if err := bind.GetDnsHandler().ZoneUpdate(domain); err != nil {
		zlog.Errorf("update domain %s failed, err: %s", domain, err.Error())
	}
}

func handleDomainDelete(domain string) {
	if err := bind.GetDnsHandler().ZoneDelete(domain); err != nil {
		zlog.Errorf("delete domain %s failed, err: %s", domain, err.Error())
	}
}
