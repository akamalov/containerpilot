package main

import (
	"fmt"
	consul "github.com/hashicorp/consul/api"
	"time"
)

const consulAddress = "consul:8500"

// ConsulProbe is a test probe for consul
type ConsulProbe interface {
	WaitForServices(service string, tag string, count int) error
}

type consulClient struct {
	Client *consul.Client
}

// NewConsulProbe creates a new ConsulProbe for testing consul
func NewConsulProbe() (ConsulProbe, error) {
	client, err := consul.NewClient(&consul.Config{
		Address: consulAddress,
		Scheme:  "http",
	})
	if err != nil {
		return nil, err
	}
	return ConsulProbe(consulClient{Client: client}), nil
}

// WaitForServices waits for the healthy services count to equal the count
// provided or it returns an error
func (c consulClient) WaitForServices(service string, tag string, count int) error {

	maxRetry := 30
	retry := 0
	var err error

	// we need to wait for Consul to start and self-elect
	for ; retry < maxRetry; retry++ {
		if retry > 0 {
			time.Sleep(1 * time.Second)
		}
		if leader, err := c.Client.Status().Leader(); err == nil && leader != "" {
			break
		}
	}
	if err != nil {
		return err
	}

	retry = 0
	for ; retry < maxRetry; retry++ {
		if retry > 0 {
			time.Sleep(1 * time.Second)
		}
		services, _, err := c.Client.Health().Service(service, tag, true, nil)
		if err == nil && len(services) == count {
			return nil
		}
	}
	if err != nil {
		return err
	}
	return fmt.Errorf("Service %s (tag:%s) count != %d", service, tag, count)
}
