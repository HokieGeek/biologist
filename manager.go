package biologist

import (
	"fmt"
)

// Manager keeps track of all Biologist instances
type Manager struct { // {{{
	biologists map[string]*Biologist
}

func (t *Manager) stringID(id []byte) string {
	return fmt.Sprintf("%x", id)
}

// Biologist returns the instalce of Biologist with the given ID
func (t *Manager) Biologist(id []byte) *Biologist {
	// TODO: validate the input
	return t.biologists[t.stringID(id)]
}

// Add keeps track of a new Biologist instance
func (t *Manager) Add(biologist *Biologist) {
	// TODO: validate the input
	t.biologists[t.stringID(biologist.ID)] = biologist
}

// Remove deletes the Biologist instance of the given ID
func (t *Manager) Remove(id []byte) {
	// TODO: validate the input
	delete(t.biologists, t.stringID(id))
}

// NewManager creates a new instance of the Biologist manager
func NewManager() *Manager {
	m := new(Manager)

	m.biologists = make(map[string]*Biologist, 0)

	return m
} // }}}
