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
	Status  Status
	Living  []life.Location
	Changes []ChangedLocation
	// TODO: checksum []byte
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

// type (t *Analysis) Checksum() [sha1.Size]byte {
// var str bytes.Buffer
// str.WriteString(strconv.Itoa(t.Generations))

// h := sha1.New()
// buf := make([]byte, sha1.Size)
// h.Write(buf)
// return h.Sum(nil)
// }

type Biologist struct { // {{{
	Id           []byte
	Life         *life.Life
	analyses     []Analysis // Each index is a generation
	stopAnalysis func()
}

func (t *Biologist) Analysis(generation int) *Analysis {
	if generation < 0 || generation >= len(t.analyses) {
		// TODO: maybe an error
		return nil
	}
	return &t.analyses[generation]
}

func (t *Biologist) analyze(generation *life.Generation) {
	var analysis Analysis

	// Record the status
	// analysis.Status =

	// Copy the living cells
	analysis.Living = make([]life.Location, len(generation.Living))
	copy(analysis.Living, generation.Living)

	// Initialize and start processing the living cells
	analysis.Changes = make([]ChangedLocation, 0)

	if generation.Num <= 0 { // Special case to reduce code duplication
		for _, loc := range generation.Living {
			analysis.Changes = append(analysis.Changes, ChangedLocation{Location: loc, Change: Born})
		}
	} else {
		// Add any new cells
		previousLiving := t.analyses[generation.Num-1].Living
		for _, newCell := range generation.Living {
			found := false
			for _, oldCell := range previousLiving {
				if oldCell.Equals(&newCell) {
					found = true
					break
				}
			}

			if !found {
				analysis.Changes = append(analysis.Changes, ChangedLocation{Location: newCell, Change: Born})
			}
		}

		// Add any cells which died
		for _, oldCell := range previousLiving {
			found := false
			for _, newCell := range generation.Living {
				if newCell.Equals(&oldCell) {
					found = true
					break
				}
			}

			if !found {
				analysis.Changes = append(analysis.Changes, ChangedLocation{Location: oldCell, Change: Died})
			}
		}

	}

	t.analyses = append(t.analyses, analysis)
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
				// nextGen := len(t.analyses)
				// gen := t.Life.Generation(nextGen)
				fmt.Printf("Generation %d\n", gen.Num)
				fmt.Println(t.Life)
				t.analyze(gen)
			}
		}
	}()
}

// func (t *Biologist) StartOLD() {
// 	updates := make(chan bool)
// 	t.stopAnalysis = t.Life.Start(updates, -1)
//
// 	go func() {
// 		for {
// 			select {
// 			case <-updates:
// 				nextGen := len(t.analyses)
// 				gen := t.Life.Generation(nextGen)
// 				fmt.Printf("Generation %d\n", gen.Num)
// 				fmt.Println(t.Life)
// 				t.analyze(gen)
// 			}
// 		}
// 	}()
// }

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
	a := new(Biologist)

	var err error
	a.Life, err = life.New("HTTP REQUEST",
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
	a.Id = uniqueId()

	// Generate first analysis (for generation 0 / the seed)
	a.analyze(&life.Generation{Living: a.Life.Seed, Num: 0})

	return a, nil
}

// vim: set foldmethod=marker:
