package biologist

import (
	"gitlab.com/hokiegeek/life"
	"testing"
	"time"
)

func TestUniqueID(t *testing.T) { // {{{
	id := uniqueID()
	if id == nil {
		t.Error("Unexpectedly got a nil unique id")
	}
} // }}}

func TestBiologistCreate(t *testing.T) { // {{{
	size := life.Dimensions{Width: 3, Height: 3}
	biologist, err := New(size, life.Blinkers, life.ConwayTester())
	if err != nil {
		t.Fatalf("Unable to create biologist: %s\n", err)
	}

	if len(biologist.Life.Seed) <= 0 {
		t.Error("Created biologist with an empty seed")
	}
}

func TestBiologistCreateError(t *testing.T) {
	size := life.Dimensions{Width: 0, Height: 0}
	_, err := New(size, life.Blinkers, life.ConwayTester())
	if err == nil {
		t.Fatal("Unexpectedly successful at creating biologist with board of 0 size")
	}
} // }}}

func TestBiologistString(t *testing.T) { // {{{
	size := life.Dimensions{Width: 3, Height: 3}
	biologist, err := New(size, life.Blinkers, life.ConwayTester())
	if err != nil {
		t.Fatalf("Unable to create biologist: %s\n", err)
	}

	if len(biologist.String()) <= 0 {
		t.Error("Biologist String function returned empty string")
	}
} // }}}

func TestBiologistStart(t *testing.T) { // {{{
	size := life.Dimensions{Width: 3, Height: 3}
	biologist, err := New(size, life.Blinkers, life.ConwayTester())
	if err != nil {
		t.Fatalf("Unable to create biologist: %s\n", err)
	}

	biologist.Start()
	waitTime := time.Millisecond * 50
	time.Sleep(waitTime)
	biologist.Stop()

	if biologist.analyses.Count() <= 0 {
		t.Fatalf("No analyses found after %s of running\n", waitTime.String())
	}
} // }}}

func TestBiologistStop(t *testing.T) { // {{{
	size := life.Dimensions{Width: 3, Height: 3}
	biologist, err := New(size, life.Blinkers, life.ConwayTester())
	if err != nil {
		t.Fatalf("Unable to create biologist: %s\n", err)
	}

	biologist.Start()
	time.Sleep(time.Millisecond * 3)
	biologist.Stop()

	time.Sleep(time.Millisecond * 1)
	stoppedCount := biologist.analyses.Count()

	time.Sleep(time.Millisecond * 10)

	waitedCount := biologist.analyses.Count()
	if stoppedCount != waitedCount {
		t.Fatalf("Analyses increased after stopped. Expected %d and got %d\n", stoppedCount, waitedCount)
	}
} // }}}

func TestBiologistAnalysis(t *testing.T) { // {{{
	size := life.Dimensions{Width: 3, Height: 3}
	biologist, err := New(size, life.Blinkers, life.ConwayTester())
	if err != nil {
		t.Fatalf("Unable to create biologist: %s\n", err)
	}

	biologist.Start()
	time.Sleep(time.Millisecond * 10)
	biologist.Stop()

	if biologist.Analysis(0) == nil {
		t.Fatal("Could not retrieve seed")
	}

	for i := biologist.analyses.Count() - 1; i >= 0; i-- {
		if biologist.Analysis(i) == nil {
			t.Fatalf("Analysis for generation %d is nil\n", i)
		}
	}
}

func TestBiologistAnalysisError(t *testing.T) {
	size := life.Dimensions{Width: 3, Height: 3}
	biologist, err := New(size, life.Blinkers, life.ConwayTester())
	if err != nil {
		t.Fatalf("Unable to create biologist: %s\n", err)
	}

	biologist.Start()
	time.Sleep(time.Millisecond * 10)
	biologist.Stop()

	if biologist.Analysis(-1) != nil {
		t.Fatal("Biologist returned to me analysis at generation -1")
	}

	// TODO: this test needs to take a Stable culture into account
	// if biologist.Analysis(biologist.analyses.Count()) != nil {
	// 	t.Fatal("Biologist returned to me analysis at generation greater than the number of generations analyzed")
	// }
} // }}}

func TestStatusString(t *testing.T) {
	var status status

	status = Seeded
	if len(status.String()) <= 0 {
		t.Error("Unexpectedly retrieved empty string from status object")
	}

	status = Active
	if len(status.String()) <= 0 {
		t.Error("Unexpectedly retrieved empty string from status object")
	}

	status = Stable
	if len(status.String()) <= 0 {
		t.Error("Unexpectedly retrieved empty string from status object")
	}

	status = Dead
	if len(status.String()) <= 0 {
		t.Error("Unexpectedly retrieved empty string from status object")
	}
}

// vim: set foldmethod=marker:
