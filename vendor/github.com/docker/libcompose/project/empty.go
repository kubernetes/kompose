package project

// EmptyService is a struct that implements Service but does nothing.
type EmptyService struct {
}

// Create implements Service.Create but does nothing.
func (e *EmptyService) Create() error {
	return nil
}

// Build implements Service.Build but does nothing.
func (e *EmptyService) Build() error {
	return nil
}

// Up implements Service.Up but does nothing.
func (e *EmptyService) Up() error {
	return nil
}

// Start implements Service.Start but does nothing.
func (e *EmptyService) Start() error {
	return nil
}

// Down implements Service.Down but does nothing.
func (e *EmptyService) Down() error {
	return nil
}

// Delete implements Service.Delete but does nothing.
func (e *EmptyService) Delete() error {
	return nil
}

// Restart implements Service.Restart but does nothing.
func (e *EmptyService) Restart() error {
	return nil
}

// Log implements Service.Log but does nothing.
func (e *EmptyService) Log() error {
	return nil
}

// Pull implements Service.Pull but does nothing.
func (e *EmptyService) Pull() error {
	return nil
}

// Kill implements Service.Kill but does nothing.
func (e *EmptyService) Kill() error {
	return nil
}

// Containers implements Service.Containers but does nothing.
func (e *EmptyService) Containers() ([]Container, error) {
	return []Container{}, nil
}

// Scale implements Service.Scale but does nothing.
func (e *EmptyService) Scale(count int) error {
	return nil
}

// Info implements Service.Info but does nothing.
func (e *EmptyService) Info(qFlag bool) (InfoSet, error) {
	return InfoSet{}, nil
}

// Pause implements Service.Pause but does nothing.
func (e *EmptyService) Pause() error {
	return nil
}

// Unpause implements Service.Pause but does nothing.
func (e *EmptyService) Unpause() error {
	return nil
}
