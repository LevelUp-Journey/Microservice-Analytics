package valueobjects

import "errors"

// ProgrammingLanguage representa los lenguajes de programación soportados
type ProgrammingLanguage string

const (
	LanguageCpp        ProgrammingLanguage = "cpp"
	LanguageJava       ProgrammingLanguage = "java"
	LanguagePython     ProgrammingLanguage = "python"
	LanguageJavaScript ProgrammingLanguage = "javascript"
	LanguageGo         ProgrammingLanguage = "go"
	LanguageRust       ProgrammingLanguage = "rust"
)

// NewProgrammingLanguage crea y valida un ProgrammingLanguage
func NewProgrammingLanguage(value string) (ProgrammingLanguage, error) {
	lang := ProgrammingLanguage(value)

	switch lang {
	case LanguageCpp, LanguageJava, LanguagePython, LanguageJavaScript, LanguageGo, LanguageRust:
		return lang, nil
	default:
		return "", errors.New("invalid programming language")
	}
}

// String implementa Stringer
func (p ProgrammingLanguage) String() string {
	return string(p)
}

// Value retorna el valor del ProgrammingLanguage
func (p ProgrammingLanguage) Value() string {
	return string(p)
}
