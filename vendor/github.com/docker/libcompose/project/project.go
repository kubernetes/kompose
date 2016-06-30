package project

import (
	"errors"
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/libcompose/logger"
	"github.com/docker/libcompose/utils"
)

// ServiceState holds the state of a service.
type ServiceState string

// State definitions
var (
	StateExecuted = ServiceState("executed")
	StateUnknown  = ServiceState("unknown")
)

// Error definitions
var (
	ErrRestart     = errors.New("Restart execution")
	ErrUnsupported = errors.New("UnsupportedOperation")
)

// Event holds project-wide event informations.
type Event struct {
	EventType   EventType
	ServiceName string
	Data        map[string]string
}

type wrapperAction func(*serviceWrapper, map[string]*serviceWrapper)
type serviceAction func(service Service) error

// NewProject create a new project with the specified context.
func NewProject(context *Context) *Project {
	p := &Project{
		context: context,
		Configs: make(map[string]*ServiceConfig),
	}

	if context.LoggerFactory == nil {
		context.LoggerFactory = &logger.NullLogger{}
	}

	context.Project = p

	p.listeners = []chan<- Event{NewDefaultListener(p)}

	return p
}

// Parse populates project information based on its context. It sets up the name,
// the composefile and the composebytes (the composefile content).
func (p *Project) Parse() error {
	err := p.context.open()
	if err != nil {
		return err
	}

	p.Name = p.context.ProjectName

	if p.context.ComposeFile == "-" {
		p.File = "."
	} else {
		p.File = p.context.ComposeFile
	}

	if p.context.ComposeBytes != nil {
		return p.Load(p.context.ComposeBytes)
	}

	return nil
}

// CreateService creates a service with the specified name based. It there
// is no config in the project for this service, it will return an error.
func (p *Project) CreateService(name string) (Service, error) {
	existing, ok := p.Configs[name]
	if !ok {
		return nil, fmt.Errorf("Failed to find service: %s", name)
	}

	// Copy because we are about to modify the environment
	config := *existing

	if p.context.EnvironmentLookup != nil {
		parsedEnv := make([]string, 0, len(config.Environment.Slice()))

		for _, env := range config.Environment.Slice() {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) > 1 && parts[1] != "" {
				parsedEnv = append(parsedEnv, env)
				continue
			} else {
				env = parts[0]
			}

			for _, value := range p.context.EnvironmentLookup.Lookup(env, name, &config) {
				parsedEnv = append(parsedEnv, value)
			}
		}

		config.Environment = NewMaporEqualSlice(parsedEnv)
	}

	return p.context.ServiceFactory.Create(p, name, &config)
}

// AddConfig adds the specified service config for the specified name.
func (p *Project) AddConfig(name string, config *ServiceConfig) error {
	p.Notify(EventServiceAdd, name, nil)

	p.Configs[name] = config
	p.reload = append(p.reload, name)

	return nil
}

// Load loads the specified byte array (the composefile content) and adds the
// service configuration to the project.
func (p *Project) Load(bytes []byte) error {
	configs := make(map[string]*ServiceConfig)
	configs, err := mergeProject(p, bytes)
	if err != nil {
		log.Errorf("Could not parse config for project %s : %v", p.Name, err)
		return err
	}

	for name, config := range configs {
		err := p.AddConfig(name, config)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Project) loadWrappers(wrappers map[string]*serviceWrapper, servicesToConstruct []string) error {
	for _, name := range servicesToConstruct {
		wrapper, err := newServiceWrapper(name, p)
		if err != nil {
			return err
		}
		wrappers[name] = wrapper
	}

	return nil
}

// Build builds the specified services (like docker build).
func (p *Project) Build(services ...string) error {
	return p.perform(EventProjectBuildStart, EventProjectBuildDone, services, wrapperAction(func(wrapper *serviceWrapper, wrappers map[string]*serviceWrapper) {
		wrapper.Do(wrappers, EventServiceBuildStart, EventServiceBuild, func(service Service) error {
			return service.Build()
		})
	}), nil)
}

// Create creates the specified services (like docker create).
func (p *Project) Create(services ...string) error {
	return p.perform(EventProjectCreateStart, EventProjectCreateDone, services, wrapperAction(func(wrapper *serviceWrapper, wrappers map[string]*serviceWrapper) {
		wrapper.Do(wrappers, EventServiceCreateStart, EventServiceCreate, func(service Service) error {
			return service.Create()
		})
	}), nil)
}

// Down stops the specified services (like docker stop).
func (p *Project) Down(services ...string) error {
	return p.perform(EventProjectDownStart, EventProjectDownDone, services, wrapperAction(func(wrapper *serviceWrapper, wrappers map[string]*serviceWrapper) {
		wrapper.Do(nil, EventServiceDownStart, EventServiceDown, func(service Service) error {
			return service.Down()
		})
	}), nil)
}

// Restart restarts the specified services (like docker restart).
func (p *Project) Restart(services ...string) error {
	return p.perform(EventProjectRestartStart, EventProjectRestartDone, services, wrapperAction(func(wrapper *serviceWrapper, wrappers map[string]*serviceWrapper) {
		wrapper.Do(wrappers, EventServiceRestartStart, EventServiceRestart, func(service Service) error {
			return service.Restart()
		})
	}), nil)
}

// Start starts the specified services (like docker start).
func (p *Project) Start(services ...string) error {
	return p.perform(EventProjectStartStart, EventProjectStartDone, services, wrapperAction(func(wrapper *serviceWrapper, wrappers map[string]*serviceWrapper) {
		wrapper.Do(wrappers, EventServiceStartStart, EventServiceStart, func(service Service) error {
			return service.Start()
		})
	}), nil)
}

// Up create and start the specified services (kinda like docker run).
func (p *Project) Up(services ...string) error {
	return p.perform(EventProjectUpStart, EventProjectUpDone, services, wrapperAction(func(wrapper *serviceWrapper, wrappers map[string]*serviceWrapper) {
		wrapper.Do(wrappers, EventServiceUpStart, EventServiceUp, func(service Service) error {
			return service.Up()
		})
	}), func(service Service) error {
		return service.Create()
	})
}

// Log aggregate and prints out the logs for the specified services.
func (p *Project) Log(services ...string) error {
	return p.forEach(services, wrapperAction(func(wrapper *serviceWrapper, wrappers map[string]*serviceWrapper) {
		wrapper.Do(nil, NoEvent, NoEvent, func(service Service) error {
			return service.Log()
		})
	}), nil)
}

// Pull pulls the specified services (like docker pull).
func (p *Project) Pull(services ...string) error {
	return p.forEach(services, wrapperAction(func(wrapper *serviceWrapper, wrappers map[string]*serviceWrapper) {
		wrapper.Do(nil, EventServicePullStart, EventServicePull, func(service Service) error {
			return service.Pull()
		})
	}), nil)
}

// Delete removes the specified services (like docker rm).
func (p *Project) Delete(services ...string) error {
	return p.perform(EventProjectDeleteStart, EventProjectDeleteDone, services, wrapperAction(func(wrapper *serviceWrapper, wrappers map[string]*serviceWrapper) {
		wrapper.Do(nil, EventServiceDeleteStart, EventServiceDelete, func(service Service) error {
			return service.Delete()
		})
	}), nil)
}

// Kill kills the specified services (like docker kill).
func (p *Project) Kill(services ...string) error {
	return p.perform(EventProjectKillStart, EventProjectKillDone, services, wrapperAction(func(wrapper *serviceWrapper, wrappers map[string]*serviceWrapper) {
		wrapper.Do(nil, EventServiceKillStart, EventServiceKill, func(service Service) error {
			return service.Kill()
		})
	}), nil)
}

// Pause pauses the specified services containers (like docker pause).
func (p *Project) Pause(services ...string) error {
	return p.perform(EventProjectPauseStart, EventProjectPauseDone, services, wrapperAction(func(wrapper *serviceWrapper, wrappers map[string]*serviceWrapper) {
		wrapper.Do(nil, EventServicePauseStart, EventServicePause, func(service Service) error {
			return service.Pause()
		})
	}), nil)
}

// Unpause pauses the specified services containers (like docker pause).
func (p *Project) Unpause(services ...string) error {
	return p.perform(EventProjectUnpauseStart, EventProjectUnpauseDone, services, wrapperAction(func(wrapper *serviceWrapper, wrappers map[string]*serviceWrapper) {
		wrapper.Do(nil, EventServiceUnpauseStart, EventServiceUnpause, func(service Service) error {
			return service.Unpause()
		})
	}), nil)
}

func (p *Project) perform(start, done EventType, services []string, action wrapperAction, cycleAction serviceAction) error {
	p.Notify(start, "", nil)

	err := p.forEach(services, action, cycleAction)

	p.Notify(done, "", nil)
	return err
}

func isSelected(wrapper *serviceWrapper, selected map[string]bool) bool {
	return len(selected) == 0 || selected[wrapper.name]
}

func (p *Project) forEach(services []string, action wrapperAction, cycleAction serviceAction) error {
	selected := make(map[string]bool)
	wrappers := make(map[string]*serviceWrapper)

	for _, s := range services {
		selected[s] = true
	}

	return p.traverse(true, selected, wrappers, action, cycleAction)
}

func (p *Project) startService(wrappers map[string]*serviceWrapper, history []string, selected, launched map[string]bool, wrapper *serviceWrapper, action wrapperAction, cycleAction serviceAction) error {
	if launched[wrapper.name] {
		return nil
	}

	launched[wrapper.name] = true
	history = append(history, wrapper.name)

	for _, dep := range wrapper.service.DependentServices() {
		target := wrappers[dep.Target]
		if target == nil {
			log.Errorf("Failed to find %s", dep.Target)
			continue
		}

		if utils.Contains(history, dep.Target) {
			cycle := strings.Join(append(history, dep.Target), "->")
			if dep.Optional {
				log.Debugf("Ignoring cycle for %s", cycle)
				wrapper.IgnoreDep(dep.Target)
				if cycleAction != nil {
					var err error
					log.Debugf("Running cycle action for %s", cycle)
					err = cycleAction(target.service)
					if err != nil {
						return err
					}
				}
			} else {
				return fmt.Errorf("Cycle detected in path %s", cycle)
			}

			continue
		}

		err := p.startService(wrappers, history, selected, launched, target, action, cycleAction)
		if err != nil {
			return err
		}
	}

	if isSelected(wrapper, selected) {
		log.Debugf("Launching action for %s", wrapper.name)
		go action(wrapper, wrappers)
	} else {
		wrapper.Ignore()
	}

	return nil
}

func (p *Project) traverse(start bool, selected map[string]bool, wrappers map[string]*serviceWrapper, action wrapperAction, cycleAction serviceAction) error {
	restart := false
	wrapperList := []string{}

	if start {
		for name := range p.Configs {
			wrapperList = append(wrapperList, name)
		}
	} else {
		for _, wrapper := range wrappers {
			if err := wrapper.Reset(); err != nil {
				return err
			}
		}
		wrapperList = p.reload
	}

	p.loadWrappers(wrappers, wrapperList)
	p.reload = []string{}

	// check service name
	for s := range selected {
		if wrappers[s] == nil {
			return errors.New("No such service: " + s)
		}
	}

	launched := map[string]bool{}

	for _, wrapper := range wrappers {
		p.startService(wrappers, []string{}, selected, launched, wrapper, action, cycleAction)
	}

	var firstError error

	for _, wrapper := range wrappers {
		if !isSelected(wrapper, selected) {
			continue
		}
		if err := wrapper.Wait(); err == ErrRestart {
			restart = true
		} else if err != nil {
			log.Errorf("Failed to start: %s : %v", wrapper.name, err)
			if firstError == nil {
				firstError = err
			}
		}
	}

	if restart {
		if p.ReloadCallback != nil {
			if err := p.ReloadCallback(); err != nil {
				log.Errorf("Failed calling callback: %v", err)
			}
		}
		return p.traverse(false, selected, wrappers, action, cycleAction)
	}
	return firstError
}

// AddListener adds the specified listener to the project.
func (p *Project) AddListener(c chan<- Event) {
	if !p.hasListeners {
		for _, l := range p.listeners {
			close(l)
		}
		p.hasListeners = true
		p.listeners = []chan<- Event{c}
	} else {
		p.listeners = append(p.listeners, c)
	}
}

// Notify notifies all project listener with the specified eventType, service name and datas.
func (p *Project) Notify(eventType EventType, serviceName string, data map[string]string) {
	if eventType == NoEvent {
		return
	}

	event := Event{
		EventType:   eventType,
		ServiceName: serviceName,
		Data:        data,
	}

	for _, l := range p.listeners {
		l <- event
	}
}
