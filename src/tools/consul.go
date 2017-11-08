package tools

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"

	consul "github.com/hashicorp/consul/api"
	"github.com/sirupsen/logrus"
)

type subConsulAgent interface {
	ServiceRegister(service *consul.AgentServiceRegistration) error
	ServiceDeregister(serviceID string) error
}

type subConsulHealth interface {
	Service(service, tag string, passingOnly bool, q *consul.QueryOptions) ([]*consul.ServiceEntry, *consul.QueryMeta, error)
}

type subConsulCatalog interface {
	Services(q *consul.QueryOptions) (map[string][]string, *consul.QueryMeta, error)
	Service(service, tag string, q *consul.QueryOptions) ([]*consul.CatalogService, *consul.QueryMeta, error)
}

//Consul permit to interact with consul api
type Consul interface {
	GetAddress() string
	Register(ip string, hostname string, httpPort int) error
	ServiceFirst(name string, prefix string) (string, error)
	DeRegister(string) error
	AllServices() ([]ConsulService, error)
}

//ConsulService represent a service registered in consul
type ConsulService struct {
	ServiceID   string
	ServiceName string
}

type consulImpl struct {
	agent   subConsulAgent
	health  subConsulHealth
	catalog subConsulCatalog
	exiter  func(int)
	Address string
}

var logConsul = logrus.WithField("logger", "tools/consul/init")

//NewConsul return a Consul
func NewConsul(consulAddress string) (Consul, error) {
	config := consul.DefaultConfig()
	config.Address = consulAddress
	c, err := consul.NewClient(config)
	if err != nil {
		return nil, err
	}

	return &consulImpl{
		agent:   c.Agent(),
		health:  c.Health(),
		catalog: c.Catalog(),
		exiter:  os.Exit,
		Address: consulAddress,
	}, nil
}

func (c *consulImpl) GetAddress() string {
	return c.Address
}

func (c *consulImpl) Register(ip string, hostname string, httpPort int) error {
	c.deRegisterOnKill(hostname)
	reg := &consul.AgentServiceRegistration{
		ID:      hostname,
		Name:    "lorhammer",
		Tags:    []string{"metrics"},
		Address: ip,
		Port:    httpPort,
	}
	return c.agent.ServiceRegister(reg)
}

func (c *consulImpl) ServiceFirst(name string, prefix string) (string, error) {
	services, _, err := c.health.Service(name, "", true, nil)
	if err != nil {
		return "", err
	}
	if len(services) <= 0 {
		return "", fmt.Errorf("No instance of service %s registered in consul", name)
	}
	return prefix + services[0].Service.Address + ":" + strconv.Itoa(services[0].Service.Port), nil
}

func (c *consulImpl) deRegisterOnKill(name string) {
	chanSignal := make(chan os.Signal, 1)
	signal.Notify(chanSignal, os.Interrupt)
	go func() {
		for range chanSignal {
			if err := c.DeRegister(name); err != nil {
				logConsul.WithField("service", name).WithError(err).Error("can't unregister")
			}
			logConsul.WithField("service", name).Info("DeRegister from consul")
			c.exiter(0)
		}
	}()
}

func (c *consulImpl) DeRegister(service string) error {
	return c.agent.ServiceDeregister(service)
}

func (c *consulImpl) AllServices() ([]ConsulService, error) {
	m, _, err := c.catalog.Services(nil)
	if err != nil {
		return nil, err
	}
	services := make([]*consul.CatalogService, 0)
	for serviceName := range m {
		s, _, err := c.catalog.Service(serviceName, "", nil)
		if err != nil {
			return nil, err
		}
		services = append(services, s...)
	}
	consulServices := make([]ConsulService, len(services))
	for i, service := range services {
		consulServices[i] = ConsulService{ServiceID: service.ServiceID, ServiceName: service.ServiceName}
	}
	return consulServices, nil
}
