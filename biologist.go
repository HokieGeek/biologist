package biologist

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"time"

	"gitlab.com/hokiegeek/life"
)

func uniqueID() []byte { // {{{
	h := sha1.New()
	buf := make([]byte, sha1.Size)
	binary.PutVarint(buf, time.Now().UnixNano())
	h.Write(buf)
	return h.Sum(nil)
} // }}}

type status int

const (
	// Seeded status means that a seed has been applied but the simulation has not started
	Seeded status = iota
	// Active means the simulation is running and is actively evolving
	Active
	// Stable applies to a simulation that has entered a cycle where all that are left are still life and oscillators
	Stable
	// Dead applies to a simulation where all of the living cells were wiped outP
	Dead
)

func (t status) String() string {
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

type changeType int // {{{

const (
	// Born applies to a cell who was born in the current generation
	Born changeType = iota
	// Died applies to a cell which died in the current generation
	Died
)

func (t changeType) String() string {
	switch t {
	case Born:
		return "Born"
	case Died:
		return "Died"
	}

	return "Unknown"
} // }}}

type changedLocation struct { // {{{
	life.Location
	Change changeType
}

func (t *changedLocation) String() string {
	var buf bytes.Buffer
	buf.WriteString("{")
	buf.WriteString(t.Change.String())
	buf.WriteString(", ")
	buf.WriteString(t.Location.String())
	buf.WriteString("}")
	return buf.String()
} // }}}

// Analysis provides the state of each analyzed generation
type Analysis struct { // {{{
	Status  status
	Living  []life.Location
	Changes []changedLocation
}

// Clone creates a deep copy of the indicated Analysis
func (t *Analysis) Clone() *Analysis {
	shadow := new(Analysis)

	shadow.Status = t.Status

	shadow.Living = make([]life.Location, len(t.Living))
	copy(shadow.Living, t.Living)

	shadow.Changes = make([]changedLocation, len(t.Changes))
	copy(shadow.Changes, t.Changes)

	return shadow
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

// Biologist runs a Life simulation and analysis each generation for emerging patterns
type Biologist struct { // {{{
	log               *log.Logger
	ID                []byte
	Life              *life.Life
	analyses          *analysisList
	stabilityDetector *stabilityDetector
	stopAnalysis      func()
}

// Analysis returns the completed analysis of the indicated generation
func (t *Biologist) Analysis(generation int) *Analysis {
	if generation < 0 {
		return nil
	}
	if generation < t.analyses.Count() {
		analysis := t.analyses.Get(generation)
		return &analysis
	} else if t.stabilityDetector.Detected {
		cycleGen := t.stabilityDetector.CycleStart + ((generation - t.stabilityDetector.CycleStart) % t.stabilityDetector.CycleLength)
		// t.log.Printf("Stable generation '%d' translated to cycle generation '%d'\n", generation, cycleGen)

		stableAnalysis := new(Analysis)
		*stableAnalysis = t.analyses.Get(cycleGen)
		stableAnalysis.Status = Stable

		return stableAnalysis
	}
	return nil
}

func (t *Biologist) calculateChanges(generation *life.Generation, previousLiving *[]life.Location) []changedLocation {
	changes := make([]changedLocation, 0)

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
			changes = append(changes, changedLocation{Location: newCell, Change: Born})
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
			changes = append(changes, changedLocation{Location: oldCell, Change: Died})
		}
	}

	return changes
}

func (t *Biologist) analyze(generation *life.Generation) status {
	var analysis Analysis

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
			analysis.Changes = append(analysis.Changes, changedLocation{Location: loc, Change: Born})
		}
	} else {
		analysis.Changes = t.calculateChanges(generation, &t.Analysis(generation.Num-1).Living)
	}

	// Detect when cycle goes stable
	if !t.stabilityDetector.Detected && t.stabilityDetector.analyze(&analysis, generation.Num) {
		t.log.Printf("Found generation %d repeats stable cycle starting at %d\n", generation.Num, t.stabilityDetector.CycleStart)
		analysis.Status = Stable
	} else {
		// Add analysis to list
		// t.log.Printf("Adding analysis of generation %d\n", generation.Num)
		t.analyses.Add(analysis)
	}

	return analysis.Status
}

// Start beings the Life simulation and analyzes each generation
func (t *Biologist) Start() {
	updates := make(chan *life.Generation)
	t.stopAnalysis = t.Life.Start(updates)

	go func() {
		for {
			select {
			case gen := <-updates:
				// t.log.Printf("Generation %d\n", gen.Num)
				// t.log.Printf("\n%s\n", t.Life)

				// if status is !Active, then stop processing updates as there is no need
				if status := t.analyze(gen); status != Active {
					t.Stop()
				}
			}
		}
	}()
}

// Stop ends the analysis and simulation
func (t *Biologist) Stop() {
	t.stopAnalysis()
}

func (t *Biologist) String() string {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("%x", t.ID))
	buf.WriteString("\n")
	buf.WriteString(t.Life.String())

	return buf.String()
} // }}}

// New creates a new biologist with the indicated life board dimension, seed and ruleset
func New(dims life.Dimensions, seed func(life.Dimensions, life.Location) []life.Location, rulesTester func(int, bool) bool) (*Biologist, error) {
	b := new(Biologist)

	var err error
	b.Life, err = life.New(
		dims,
		life.NeighborsAll,
		seed,
		rulesTester,
		life.SimultaneousProcessor)
	if err != nil {
		log.Printf("ERROR: %s\n", err)
		return nil, err
	}

	b.ID = uniqueID()
	b.log = log.New(os.Stdout, fmt.Sprintf("[biologist-%x] ", b.ID), 0)

	b.analyses = newAnalysisList()
	b.stabilityDetector = newStabilityDetector()

	// Generate first analysis (for generation 0 / the seed)
	b.analyze(&life.Generation{Living: b.Life.Seed, Num: 0})

	return b, nil
}

// vim: set foldmethod=marker:
