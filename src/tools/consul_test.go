package tools

import (
	"errors"
	"os"
	"testing"

	consul "github.com/hashicorp/consul/api"
)

const address = "127.0.0.1:8500"

func buildValidConsul(t *testing.T, agent subConsulAgent, health subConsulHealth, catalog subConsulCatalog) Consul {
	c, err := NewConsul(address)
	if c == nil {
		t.Fatal("valid consul address should return a client")
	}
	if err != nil {
		t.Fatal("valid consul address should not throw an error")
	}
	c.(*consulImpl).agent = agent
	c.(*consulImpl).health = health
	c.(*consulImpl).catalog = catalog
	c.(*consulImpl).exiter = func(code int) {}
	return c
}

func TestNewConsul(t *testing.T) {
	c := buildValidConsul(t, nil, nil, nil)
	if c.GetAddress() != address {
		t.Fatal("consul must return same address has constructor pass")
	}
}

func TestNewConsulError(t *testing.T) {
	c, err := NewConsul("none://none")
	if c != nil {
		t.Fatal("bad consul address should not return a client")
	}
	if err == nil {
		t.Fatal("bad consul address should throw an error")
	}
}

type fakeSubConsulAgent struct {
	registerError   error
	onRegister      chan *consul.AgentServiceRegistration
	deRegisterError error
	onDeRegister    chan string
}

func (f fakeSubConsulAgent) ServiceRegister(service *consul.AgentServiceRegistration) error {
	if f.onRegister != nil {
		go func() {
			f.onRegister <- service
		}()
	}
	return f.registerError
}
func (f fakeSubConsulAgent) ServiceDeregister(serviceID string) error {
	if f.onDeRegister != nil {
		go func() {
			f.onDeRegister <- serviceID
		}()
	}
	return f.deRegisterError
}

func TestConsulImpl_Register(t *testing.T) {
	onRegister := make(chan *consul.AgentServiceRegistration)
	defer close(onRegister)

	c := buildValidConsul(t, fakeSubConsulAgent{onRegister: onRegister}, nil, nil)

	if err := c.Register("ip", "hostname", 1); err != nil {
		t.Fatal("Valid register should not return err")
	}

	service := <-onRegister
	if service.Address != "ip" {
		t.Fatal("Consul Register should be called with ip")
	}
}

func TestConsulImpl_RegisterKillDeRegister(t *testing.T) {
	onDeRegister := make(chan string)
	defer close(onDeRegister)

	c := buildValidConsul(t, fakeSubConsulAgent{onDeRegister: onDeRegister}, nil, nil)

	if err := c.Register("ip", "hostname", 1); err != nil {
		t.Fatal("Valid register should not return err")
	}

	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Signal(os.Interrupt); err != nil {
		t.Fatal(err)
	}

	idService := <-onDeRegister
	if idService != "hostname" {
		t.Fatal("Deregister should be called with id")
	}
}

func TestConsulImpl_RegisterError(t *testing.T) {
	onRegister := make(chan *consul.AgentServiceRegistration)
	defer close(onRegister)
	onDeRegister := make(chan string)
	defer close(onDeRegister)

	agent := fakeSubConsulAgent{registerError: errors.New("error")}
	c := buildValidConsul(t, agent, nil, nil)

	if err := c.Register("ip", "hostname", 1); err == nil {
		t.Fatal("Invalid register should return err")
	}
}

type fakeSubConsulHealth struct {
	entries      []*consul.ServiceEntry
	serviceError error
}

func (f fakeSubConsulHealth) Service(service, tag string, passingOnly bool, q *consul.QueryOptions) ([]*consul.ServiceEntry, *consul.QueryMeta, error) {
	return f.entries, nil, f.serviceError
}

func TestConsulImpl_ServiceFirst(t *testing.T) {
	entries := make([]*consul.ServiceEntry, 1)
	entries[0] = &consul.ServiceEntry{
		Service: &consul.AgentService{Address: "/", Port: 1},
	}

	health := fakeSubConsulHealth{entries: entries}
	c := buildValidConsul(t, nil, health, nil)

	serviceAddress, err := c.ServiceFirst("", "")
	if err != nil {
		t.Fatal(err)
		t.Fatal("Valid first service should not throw err")
	}
	if serviceAddress != "/:1" {
		t.Fatal("First service should return the complete address of first service")
	}
}

func TestConsulImpl_ServiceFirstEmpty(t *testing.T) {
	health := fakeSubConsulHealth{entries: make([]*consul.ServiceEntry, 0)}
	c := buildValidConsul(t, nil, health, nil)

	serviceAddress, err := c.ServiceFirst("", "")
	if err == nil {
		t.Fatal("Empty first service should throw err")
	}
	if serviceAddress != "" {
		t.Fatal("Empty First service should not return the complete address")
	}
}

func TestConsulImpl_ServiceFirstMultiple(t *testing.T) {
	entries := make([]*consul.ServiceEntry, 2)
	entries[0] = &consul.ServiceEntry{
		Service: &consul.AgentService{Address: "/", Port: 1},
	}
	entries[1] = &consul.ServiceEntry{
		Service: &consul.AgentService{Address: "|", Port: 2},
	}

	health := fakeSubConsulHealth{entries: entries}
	c := buildValidConsul(t, nil, health, nil)

	serviceAddress, err := c.ServiceFirst("", "")
	if err != nil {
		t.Fatal("Valid first service with multiple services should not throw err")
	}
	if serviceAddress != "/:1" {
		t.Fatal("First service with multiple services should return the complete address of first service")
	}
}

func TestConsulImpl_ServiceFirstError(t *testing.T) {
	health := fakeSubConsulHealth{serviceError: errors.New("error")}
	c := buildValidConsul(t, nil, health, nil)

	serviceAddress, err := c.ServiceFirst("", "")
	if err == nil {
		t.Fatal("Error first service should throw err")
	}
	if serviceAddress != "" {
		t.Fatal("Error First service should not return the address of first service")
	}
}

type fakeSubConsulCatalog struct {
	services      map[string][]string
	infos         []*consul.CatalogService
	serviceError  error
	servicesError error
}

func (f fakeSubConsulCatalog) Services(q *consul.QueryOptions) (map[string][]string, *consul.QueryMeta, error) {
	return f.services, nil, f.servicesError
}
func (f fakeSubConsulCatalog) Service(service, tag string, q *consul.QueryOptions) ([]*consul.CatalogService, *consul.QueryMeta, error) {
	return f.infos, nil, f.serviceError
}

func TestConsulImpl_AllServices(t *testing.T) {
	fakeService := make(map[string][]string, 1)
	fakeService["serviceID"] = make([]string, 1)
	fakeService["serviceID"][0] = "serviceName"

	fakeServiceInfos := make([]*consul.CatalogService, 1)
	fakeServiceInfos[0] = &consul.CatalogService{ServiceID: "id", ServiceName: "name"}

	catalog := fakeSubConsulCatalog{services: fakeService, infos: fakeServiceInfos}
	c := buildValidConsul(t, nil, nil, catalog)

	services, err := c.AllServices()
	if err != nil {
		t.Fatal("AllServices should not throw err")
	}
	if len(services) != 1 {
		t.Fatal("AllServices services should return 1 addresse")
	}
	if services[0].ServiceID != "id" || services[0].ServiceName != "name" {
		t.Fatal("Should return good infos on services")
	}
}

func TestConsulImpl_AllServicesEmpty(t *testing.T) {
	catalog := fakeSubConsulCatalog{}
	c := buildValidConsul(t, nil, nil, catalog)

	services, err := c.AllServices()
	if err != nil {
		t.Fatal("Empty services should not throw err")
	}
	if len(services) != 0 {
		t.Fatal("Empty services should return empty addresses")
	}
}

func TestConsulImpl_AllServicesErrorServices(t *testing.T) {
	catalog := fakeSubConsulCatalog{servicesError: errors.New("error")}
	c := buildValidConsul(t, nil, nil, catalog)

	services, err := c.AllServices()
	if err == nil {
		t.Fatal("Error services should throw err")
	}
	if services != nil {
		t.Fatal("Error services should return nil addresses")
	}
}

func TestConsulImpl_AllServicesErrorService(t *testing.T) {
	fakeService := make(map[string][]string, 1)
	fakeService["serviceID"] = make([]string, 1)
	fakeService["serviceID"][0] = "serviceName"
	catalog := fakeSubConsulCatalog{services: fakeService, serviceError: errors.New("error")}
	c := buildValidConsul(t, nil, nil, catalog)

	services, err := c.AllServices()
	if err == nil {
		t.Fatal("Error service should throw err")
	}
	if services != nil {
		t.Fatal("Error service should return nil addresses")
	}
}
