/*
Package alias defines the core domain entity for an alias.
*/
package alias

/*
Alias represents a suggested alias, consisting of a short name and the
full command it expands to. This is a core domain entity.
*/
type Alias struct {
	Command string `yaml:"command"`
	Name    string `yaml:"alias"`
}
