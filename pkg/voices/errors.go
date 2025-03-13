package voices

import "errors"

// Error definitions for the voices package.
var (
	// ErrNotCreated is returned when VoicesManager.Find() is called before VoicesManager.Create().
	ErrNotCreated = errors.New("VoicesManager.Find() called before VoicesManager.Create()")
)
