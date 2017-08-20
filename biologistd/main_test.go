package main

import (
	"gitlab.com/hokiegeek/biologist"
	"gitlab.com/hokiegeek/life"
	"testing"
)

func TestNewCreateAnalysisResponse(t *testing.T) {
	size := life.Dimensions{Width: 3, Height: 3}
	biologist, err := biologist.New(size, life.Blinkers, life.ConwayTester())
	if err != nil {
		t.Fatalf("Unable to create biologist: %s\n", err)
	}

	resp := newCreateAnalysisResponse(biologist)

	if !resp.Dims.Equals(&size) {
		t.Fatalf("Expected size %s but received %s\n", size.String(), resp.Dims.String())
	}
}

/*
func TestNewBiologistUpdateResponse(t *testing.T) {
	size := life.Dimensions{Width: 3, Height: 3}
	biologist, err := life.New(size, life.Blinkers, life.ConwayTester())
	if err != nil {
		t.Fatalf("Unable to create biologist: %s\n", err)
	}

	resp := NewBiologistUpdateResponse(biologist, 0, 1)
}
*/

// vim: set foldmethod=marker:
