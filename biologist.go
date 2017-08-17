package biologist

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"gitlab.com/hokiegeek/life"
	"log"
	"os"
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
	Status  Status
	Living  []life.Location
	Changes []ChangedLocation
}

func (t *Analysis) Clone() *Analysis {
	shadow := new(Analysis)

	shadow.Status = t.Status

	shadow.Living = make([]life.Location, len(t.Living))
	copy(shadow.Living, t.Living)

	shadow.Changes = make([]ChangedLocation, len(t.Changes))
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

type Biologist struct { // {{{
	log               *log.Logger
	Id                []byte
	Life              *life.Life
	analyses          []Analysis // Each index is a generation
	stabilityDetector *stabilityDetector
	stopAnalysis      func()
}

func (t *Biologist) Analysis(generation int) *Analysis {
	if generation >= 0 && generation < len(t.analyses) {
		return &t.analyses[generation]
	} else if t.stabilityDetector.Detected {
		cycleGen := t.stabilityDetector.CycleStart + ((generation - t.stabilityDetector.CycleStart) % t.stabilityDetector.CycleLength)
		// log.Printf("Received create request: %s\n", req.String())
		t.log.Printf("Stable generation '%d' translated to cycle generation '%d'\n", generation, cycleGen)

		stableAnalysis := t.Analysis(cycleGen).Clone()
		stableAnalysis.Status = Stable

		return stableAnalysis
	}
	return nil
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
		analysis.Changes = t.calculateChanges(generation, &t.Analysis(generation.Num-1).Living)
	}

	// Detect when cycle goes stable
	if t.stabilityDetector.analyze(&analysis, generation.Num) {
		analysis.Status = Stable
	} else {
		// Add analysis to list
		t.analyses = append(t.analyses, analysis)
	}

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
	b := new(Biologist)
	b.log = log.New(os.Stdout, "[biologist] ", 0)

	var err error
	b.Life, err = life.New(
		dims,
		life.NEIGHBORS_ALL,
		pattern,
		rulesTester,
		life.SimultaneousProcessor)
	if err != nil {
		log.Printf("ERROR: %s\n", err)
		return nil, err
	}

	b.Id = uniqueId()

	b.stabilityDetector = newStabilityDetector()

	// Generate first analysis (for generation 0 / the seed)
	b.analyze(&life.Generation{Living: b.Life.Seed, Num: 0})

	return b, nil
}

// vim: set foldmethod=marker:
