package biologist

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/hokiegeek/life"
	"sort"
	"strconv"
	"time"
)

func uniqueId() []byte { // {{{
	h := sha1.New()
	buf := make([]byte, sha1.Size)
	binary.PutVarint(buf, time.Now().UnixNano())
	h.Write(buf)
	return h.Sum(nil)
} // }}}

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
	Status   life.Status
	Living   []life.Location
	Changes  []ChangedLocation
	checksum []byte
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

type lifeStableCycle struct { // {{{
	analyzer      *Analyzer
	startIndex    int
	cycleLength   int
	cycleAnalyses []*Analysis
}

func (t *lifeStableCycle) createCycleAnalysis(cyclePos int) *Analysis {
	analysis := t.analyzer.Analysis(cyclePos)

	// Copy cycleAnalysis
	cycleAnalysis := new(Analysis)
	cycleAnalysis.Status = life.Stable

	fmt.Println("~~ Copying Living ~~")
	cycleAnalysis.Living = make([]life.Location, len(analysis.Living))
	copy(cycleAnalysis.Living, analysis.Living)

	fmt.Println("~~ Copying Changes ~~")
	cycleAnalysis.Changes = make([]ChangedLocation, len(analysis.Changes))
	copy(cycleAnalysis.Changes, analysis.Changes)

	fmt.Println("~~ Copying checksum ~~")
	cycleAnalysis.checksum = analysis.checksum

	return cycleAnalysis
}

func (t *lifeStableCycle) Analysis(generation int) *Analysis {
	fmt.Printf("Analysis(%d)\n", generation)
	// cyclePos = gen_past_active_end % cycleLength
	// gen_past_active_end = gen - active_len

	cyclePos := (generation - t.analyzer.NumAnalyses()) % t.cycleLength

	if t.cycleAnalyses[cyclePos] == nil {
		t.cycleAnalyses[cyclePos] = t.createCycleAnalysis(cyclePos)
	}

	return t.cycleAnalyses[cyclePos]
}

func newLifeStableCycle(analyzer *Analyzer, startIndex int) (*lifeStableCycle, error) {
	c := new(lifeStableCycle)

	fmt.Printf(">> Found end cycle at: %d\n", startIndex)
	c.analyzer = analyzer
	c.startIndex = startIndex
	c.cycleLength = c.analyzer.NumAnalyses() - c.startIndex
	c.cycleAnalyses = make([]*Analysis, c.cycleLength)

	return c, nil
} // }}}

type Analyzer struct { // {{{
	Id                []byte
	Life              *life.Life
	analyses          []Analysis // Each index is a generation
	analysesChecksums map[string]int
	cycle             *lifeStableCycle
	analyzing         bool
	stopAnalysis      func()
}

func (t *Analyzer) Analysis(generation int) *Analysis {
	fmt.Printf("analyzed: Analysis(%d) : %d\n", generation, len(t.analyses))
	if generation < 0 {
		return nil
	}

	if generation >= len(t.analyses) {
		if t.analyzing {
			fmt.Println("analyzed: TODO: ERROR")
			// TODO: return an error
		} else {
			fmt.Println("analyzed: TODO: Returning cycle analysis")
			// return t.cycle.Analysis(generation)
		}
		return nil // TODO: remove
	}
	return &t.analyses[generation]
}

func checksum(cells []life.Location) []byte {
	/*
		var str bytes.Buffer
		fmt.Printf("checksum(")
		for _, loc := range cells {
			fmt.Printf(loc.String())
			str.WriteString(strconv.Itoa(loc.X))
			str.WriteString(strconv.Itoa(loc.Y))
		}
		fmt.Println()
	*/

	fmt.Printf("checksum(")
	locations := make(map[int]int)
	var sorted []int
	for _, loc := range cells {
		sorted = append(sorted, loc.X)
		locations[loc.X] = loc.Y
	}
	sort.Ints(sorted)

	var str bytes.Buffer
	for _, x := range sorted {
		fmt.Printf("%d,%d ", x, locations[x])
		str.WriteString(strconv.Itoa(x))
		str.WriteString(strconv.Itoa(locations[x]))
	}
	fmt.Println(")")

	h := sha1.New()
	h.Write([]byte(str.String()))
	return h.Sum(nil)
}

func (t *Analyzer) analyze(cells []life.Location, generation int) {
	fmt.Printf("analyze(..., %d)\n", generation)
	var analysis Analysis

	// Record the status
	// analysis.Status =

	// Copy the living cells
	analysis.Living = make([]life.Location, len(cells))
	copy(analysis.Living, cells)

	// Record the dead status, if applicable
	if len(analysis.Living) <= 0 {
		analysis.Status = life.Dead
	}

	analysis.checksum = checksum(analysis.Living)
	checksumStr := hex.EncodeToString(analysis.checksum)
	fmt.Printf("Analyzing: %s\n", checksumStr)
	if gen, exists := t.analysesChecksums[checksumStr]; exists {
		analysis.Status = life.Stable // TODO: remove
		// fmt.Printf(">>>> Found cycle start: %d\n", gen)
		t.cycle, _ = newLifeStableCycle(t, gen)
		t.Stop()
		fmt.Println("analyzed: Stopped analysis")
	} else {
		analysis.Status = life.Active
		if t.analysesChecksums == nil {
			t.analysesChecksums = make(map[string]int)
		}
		t.analysesChecksums[checksumStr] = len(t.analyses)
		t.analyses = append(t.analyses, analysis)
	}

	fmt.Printf("analyzed: Generation: %d  Status: %s\n", generation, analysis.Status)
}

func (t *Analyzer) Changes(olderGeneration int, newerGeneration int) []ChangedLocation { // {{{
	changes := make([]ChangedLocation, 0)

	olderLiving := t.analyses[olderGeneration].Living
	newerLiving := t.analyses[newerGeneration].Living

	// Add any new cells
	for _, newCell := range newerLiving {
		found := false
		for _, oldCell := range olderLiving {
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
	for _, oldCell := range olderLiving {
		found := false
		for _, newCell := range newerLiving {
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
} // }}}

func (t *Analyzer) NumAnalyses() int {
	return len(t.analyses)
}

func (t *Analyzer) Start() {
	t.analyzing = true
	updates := make(chan bool)
	// t.Status = life.Active
	t.stopAnalysis = t.Life.Start(updates, -1)

	go func() {
		for {
			select {
			case <-updates:
				nextGen := len(t.analyses)
				gen := t.Life.Generation(nextGen)
				fmt.Println("=======================================================")
				fmt.Printf("Generation %d\n", gen.Num)
				fmt.Println(t.Life)
				t.analyze(gen.Living, gen.Num)
			}
		}
	}()
}

func (t *Analyzer) Stop() {
	t.analyzing = false
	t.stopAnalysis()
}

func (t *Analyzer) String() string {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("%x", t.Id))
	buf.WriteString("\n")
	buf.WriteString(t.Life.String())

	return buf.String()
} // }}}

func NewAnalyzer(dims life.Dimensions, pattern func(life.Dimensions, life.Location) []life.Location, rulesTester func(int, bool) bool) (*Analyzer, error) {
	// fmt.Printf("NewAnalyzer: %v\n", pattern(dims, Location{X: 0, Y: 0}))
	a := new(Analyzer)

	var err error
	a.Life, err = life.New("Analyzer",
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
	a.analyze(a.Life.Seed, 0)

	// Initialize the checksums map
	a.analysesChecksums = make(map[string]int)

	return a, nil
}

// vim: set foldmethod=marker nofoldenable:
