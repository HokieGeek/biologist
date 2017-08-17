package biologist

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	// "fmt"
	"log"
	"os"
	"sort"
	"strconv"

	"gitlab.com/hokiegeek/life"
)

func checksum(cells []life.Location) []byte {
	// fmt.Printf("checksum(")
	locations := make(map[int]int)
	var sorted []int
	for _, loc := range cells {
		sorted = append(sorted, loc.X)
		locations[loc.X] = loc.Y
	}
	sort.Ints(sorted)

	var str bytes.Buffer
	for _, x := range sorted {
		// fmt.Printf("%d,%d ", x, locations[x]) // Debug
		str.WriteString(strconv.Itoa(x))
		str.WriteString(strconv.Itoa(locations[x]))
	}
	// fmt.Println(")")

	h := sha1.New()
	h.Write([]byte(str.String()))
	return h.Sum(nil)
}

type stabilityDetector struct { // {{{
	log               *log.Logger
	analysesChecksums map[string]int
	Detected          bool
	CycleStart        int
	CycleLength       int
}

func (s *stabilityDetector) analyze(analysis *Analysis, generation int) bool {

	checksum := checksum(analysis.Living)
	checksumStr := hex.EncodeToString(checksum)

	if gen, exists := s.analysesChecksums[checksumStr]; exists {
		s.Detected = true
		s.CycleStart = gen
		s.CycleLength = generation - gen
		s.log.Printf("Found cycle start: %d, len: %d\n", s.CycleStart, s.CycleLength)
	} else {
		s.analysesChecksums[checksumStr] = generation
	}

	return s.Detected
}

func newStabilityDetector() *stabilityDetector {
	s := new(stabilityDetector)
	s.log = log.New(os.Stdout, "[stabilityDetector] ", 0)

	s.analysesChecksums = make(map[string]int)
	s.Detected = false
	s.CycleStart = -1
	s.CycleLength = 0

	return s
}
