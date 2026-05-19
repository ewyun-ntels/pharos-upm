package permissions

// CATV Extension roles (CRUD pattern)
// Core modules must NOT import this package.
const (
	Read   = "extension:catv:read"   // View any CATV data (GET)
	Create = "extension:catv:create" // Submit control commands (POST)
	Update = "extension:catv:update" // Modify settings (PUT)
	Delete = "extension:catv:delete" // Remove data (DELETE)
)
