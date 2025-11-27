package valueobjects

import (
	"errors"
	"strings"
)

// Provider representa el proveedor de autenticación (google, facebook, github, etc.)
type Provider struct {
	value string
}

// Proveedores válidos
var validProviders = map[string]bool{
	"google":   true,
	"facebook": true,
	"github":   true,
	"twitter":  true,
	"local":    true,
	"apple":    true,
	"microsoft": true,
}

// NewProvider crea un nuevo Provider validado
func NewProvider(value string) (Provider, error) {
	if value == "" {
		return Provider{}, errors.New("provider cannot be empty")
	}

	// Normalizar a lowercase
	normalizedValue := strings.ToLower(strings.TrimSpace(value))

	if !validProviders[normalizedValue] {
		return Provider{}, errors.New("invalid provider: " + value)
	}

	return Provider{value: normalizedValue}, nil
}

// Value retorna el valor del Provider
func (p Provider) Value() string {
	return p.value
}

// String implementa fmt.Stringer
func (p Provider) String() string {
	return p.value
}

// Equals compara dos Providers
func (p Provider) Equals(other Provider) bool {
	return p.value == other.value
}

// IsOAuth indica si el provider es de tipo OAuth
func (p Provider) IsOAuth() bool {
	return p.value != "local"
}
