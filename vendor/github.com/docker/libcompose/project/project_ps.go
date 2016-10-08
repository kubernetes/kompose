package project

import (
	"golang.org/x/net/context"
)

// Ps list containers for the specified services.
func (p *Project) Ps(ctx context.Context, services ...string) (InfoSet, error) {
	allInfo := InfoSet{}
	for _, name := range p.ServiceConfigs.Keys() {
		service, err := p.CreateService(name)
		if err != nil {
			return nil, err
		}

		info, err := service.Info(ctx)
		if err != nil {
			return nil, err
		}

		allInfo = append(allInfo, info...)
	}
	return allInfo, nil
}
