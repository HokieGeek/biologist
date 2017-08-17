package biologist

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"

	"gitlab.com/hokiegeek/life"
)

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
		fmt.Printf("%d,%d ", x, locations[x]) // Debug
		str.WriteString(strconv.Itoa(x))
		str.WriteString(strconv.Itoa(locations[x]))
	}
	fmt.Println(")")

	h := sha1.New()
	h.Write([]byte(str.String()))
	return h.Sum(nil)
}

type stabilityDetector struct { // {{{
	analysesChecksums map[string]int
	CycleStart        int
	CycleEnd          int
}

func (s *stabilityDetector) analyze(analysis *Analysis, generation int) bool {

	checksum := checksum(analysis.Living)
	checksumStr := hex.EncodeToString(checksum)

	if gen, exists := s.analysesChecksums[checksumStr]; exists {
		fmt.Printf("stabilityDetector: Found cycle start: %d\n", gen)
		s.CycleStart = gen
		s.CycleEnd = generation - 1
		return true
	} else {
		s.analysesChecksums[checksumStr] = generation
	}

	return false
}

func newStabilityDetector() *stabilityDetector {
	s := new(stabilityDetector)

	s.analysesChecksums = make(map[string]int)
	s.CycleStart = -1
	s.CycleEnd = -1

	return s
}
