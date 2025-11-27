package valueobjects

import (
	"errors"
	"regexp"
	"strings"
)

// Email representa una dirección de correo electrónico válida
type Email struct {
	value string
}

// emailRegex es la expresión regular para validar emails
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// NewEmail crea un nuevo Email validado
func NewEmail(value string) (Email, error) {
	if value == "" {
		return Email{}, errors.New("email cannot be empty")
	}

	// Normalizar email a lowercase
	normalizedValue := strings.ToLower(strings.TrimSpace(value))

	// Validar formato de email
	if !emailRegex.MatchString(normalizedValue) {
		return Email{}, errors.New("invalid email format: " + value)
	}

	return Email{value: normalizedValue}, nil
}

// Value retorna el valor del Email
func (e Email) Value() string {
	return e.value
}

// String implementa fmt.Stringer
func (e Email) String() string {
	return e.value
}

// Equals compara dos Emails
func (e Email) Equals(other Email) bool {
	return e.value == other.value
}

// Domain extrae el dominio del email (ej: gmail.com)
func (e Email) Domain() string {
	parts := strings.Split(e.value, "@")
	if len(parts) == 2 {
		return parts[1]
	}
	return ""
}
