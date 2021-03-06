package internal

import log "github.com/Sirupsen/logrus"

type Loadbalancer interface {
	ApplyConfiguration(data *EventMessageWithConf) (int, error)
	Stop() error
	Delete() error
	Fake() bool
	RestartKilledHaproxy() error
}

type LoadbalancerFactory struct {
	Fake       string
	Properties *Config
}

func NewLoadbalancerFactory() *LoadbalancerFactory {
	return &LoadbalancerFactory{
		Fake: "none",
	}
}

func (factory *LoadbalancerFactory) CreateHaproxy(context Context) Loadbalancer {
	if factory.Fake == "drunk" {
		log.Debug("mode drunk")
		return &DrunkHaproxy{
			context: context,
		}
	} else if factory.Fake == "yesman" {
		log.Debug("mode yesman")
		return &YesmanHaproxy{}
	} else {
		log.Debug("process haproxy on normal mode")
		return NewHaproxy(factory.Properties, &context)
	}
}
