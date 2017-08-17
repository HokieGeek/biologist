package biologist

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"gitlab.com/hokiegeek/life"
	"time"
)

func uniqueId() []byte { // {{{
	h := sha1.New()
	buf := make([]byte, sha1.Size)
	binary.PutVarint(buf, time.Now().UnixNano())
	h.Write(buf)
	return h.Sum(nil)
} // }}}

type Status int

const (
	Seeded Status = iota
	Active
	Stable
	Dead
)

func (t Status) String() string {
	switch t {
	case Seeded:
		return "Seeded"
	case Active:
		return "Active"
	case Stable:
		return "Stable"
	case Dead:
		return "Dead"
	}

	return "Unknown"
}

type ChangeType int // {{{

const (
	Born ChangeType = iota
	Died
)

func (t ChangeType) String() string {
	switch t {
	case Born:
		return "Born"
	case Died:
		return "Died"
	}

	return "Unknown"
} // }}}

type ChangedLocation struct { // {{{
	life.Location
	Change ChangeType
	// PatternGroup ...
	// Classificaiton ...
}

func (t *ChangedLocation) String() string {
	var buf bytes.Buffer
	buf.WriteString("{")
	buf.WriteString(t.Change.String())
	buf.WriteString(", ")
	buf.WriteString(t.Location.String())
	buf.WriteString("}")
	return buf.String()
} // }}}

type Analysis struct { // {{{
	Status     Status
	Generation int
	Living     []life.Location
	Changes    []ChangedLocation
	// checksum []byte
}

func (t *Analysis) String() string {
	var buf bytes.Buffer
	buf.WriteString("Analysis {")
	buf.WriteString("\n\tStatus = ")
	buf.WriteString(t.Status.String())
	buf.WriteString("\n\tLiving = {")
	for _, living := range t.Living {
		buf.WriteString("\n\t\t")
		buf.WriteString(living.String())
	}
	buf.WriteString("\n\t}")
	buf.WriteString("\n\tChanged = {")
	for _, change := range t.Changes {
		buf.WriteString("\n\t\t")
		buf.WriteString(change.String())
	}
	buf.WriteString("\n\t}")
	buf.WriteString("\n}")
	return buf.String()
} // }}}

type Biologist struct { // {{{
	Id                []byte
	Life              *life.Life
	analyses          []Analysis // Each index is a generation
	stabilityDetector *stabilityDetector
	stopAnalysis      func()
}

func (t *Biologist) Analysis(generation int) *Analysis {
	if generation < 0 || generation >= len(t.analyses) {
		// TODO: maybe an error
		return nil
	}
	return &t.analyses[generation]
}

func (t *Biologist) calculateChanges(generation *life.Generation, previousLiving *[]life.Location) []ChangedLocation {
	changes := make([]ChangedLocation, 0)

	// Add any new cells
	for _, newCell := range generation.Living {
		found := false
		for _, oldCell := range *previousLiving {
			if oldCell.Equals(&newCell) {
				found = true
				break
			}
		}

		if !found {
			changes = append(changes, ChangedLocation{Location: newCell, Change: Born})
		}
	}

	// Add any cells which died
	for _, oldCell := range *previousLiving {
		found := false
		for _, newCell := range generation.Living {
			if newCell.Equals(&oldCell) {
				found = true
				break
			}
		}

		if !found {
			changes = append(changes, ChangedLocation{Location: oldCell, Change: Died})
		}
	}

	return changes
}

func (t *Biologist) analyze(generation *life.Generation) Status {
	var analysis Analysis

	analysis.Generation = generation.Num

	// Assume active status
	analysis.Status = Active

	// Copy the living cells
	analysis.Living = make([]life.Location, len(generation.Living))
	copy(analysis.Living, generation.Living)

	if len(analysis.Living) == 0 {
		analysis.Status = Dead
	}

	// Initialize and start processing the living cells
	if generation.Num <= 0 { // Special case to reduce code duplication
		for _, loc := range generation.Living {
			analysis.Changes = append(analysis.Changes, ChangedLocation{Location: loc, Change: Born})
		}
	} else {
		analysis.Changes = t.calculateChanges(generation, &t.analyses[generation.Num-1].Living)
	}

	// Detect when cycle goes stable
	if t.stabilityDetector.analyze(&analysis) {
		analysis.Status = Stable
	}

	// Add analysis to list
	t.analyses = append(t.analyses, analysis)
	return analysis.Status
}

func (t *Biologist) NumAnalyses() int {
	return len(t.analyses)
}

func (t *Biologist) Start() {
	updates := make(chan *life.Generation)
	t.stopAnalysis = t.Life.Start(updates)

	go func() {
		for {
			select {
			case gen := <-updates:
				fmt.Printf("Generation %d\n", gen.Num)
				fmt.Println(t.Life)

				// if status is !Active, then stop processing updates as there is no need
				if t.analyze(gen) != Active {
					t.Stop()
				}
			}
		}
	}()
}

func (t *Biologist) Stop() {
	t.stopAnalysis()
}

func (t *Biologist) String() string {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("%x", t.Id))
	buf.WriteString("\n")
	buf.WriteString(t.Life.String())

	return buf.String()
} // }}}

func New(dims life.Dimensions, pattern func(life.Dimensions, life.Location) []life.Location, rulesTester func(int, bool) bool) (*Biologist, error) {
	// fmt.Printf("NewBiologist: %v\n", pattern(dims, Location{X: 0, Y: 0}))
	b := new(Biologist)

	var err error
	b.Life, err = life.New(
		dims,
		life.NEIGHBORS_ALL,
		pattern,
		rulesTester,
		life.SimultaneousProcessor)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return nil, err
	}

	// fmt.Println("Creating unique id")
	b.Id = uniqueId()

	b.stabilityDetector = newStabilityDetector()

	// Generate first analysis (for generation 0 / the seed)
	b.analyze(&life.Generation{Living: b.Life.Seed, Num: 0})

	return b, nil
}

// vim: set foldmethod=marker:
