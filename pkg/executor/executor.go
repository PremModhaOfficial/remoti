package executor

// Executor defines the interface for interacting with the OS-level input system.
type Executor interface {
	// Type types a raw string out.
	Type(text string) error
	// Combo presses a combination of keys simultaneously and then releases them.
	Combo(keys ...string) error
	// KeyDown presses and holds a key.
	KeyDown(key string) error
	// KeyUp releases a held key.
	KeyUp(key string) error
	// Reset releases all currently held keys.
	Reset() error
	// Close cleans up the executor resources.
	Close() error
}
