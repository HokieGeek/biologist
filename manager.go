package biologist

import (
	"fmt"
)

type Manager struct { // {{{
	biologists map[string]*Biologist
}

func (t *Manager) stringId(id []byte) string {
	return fmt.Sprintf("%x", id)
}

func (t *Manager) Biologist(id []byte) *Biologist {
	// TODO: validate the input
	return t.biologists[t.stringId(id)]
}

func (t *Manager) Add(biologist *Biologist) {
	// TODO: validate the input
	t.biologists[t.stringId(biologist.Id)] = biologist
}

func (t *Manager) Remove(id []byte) {
	// TODO: validate the input
	delete(t.biologists, t.stringId(id))
}

func NewManager() *Manager {
	m := new(Manager)

	m.biologists = make(map[string]*Biologist, 0)

	return m
} // }}}
