package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"gitlab.com/hokiegeek/biologist"
	"gitlab.com/hokiegeek/life"
)

/////////////////////////////////// CREATE ANALYSIS ///////////////////////////////////

// CreateAnalysisResponse encapsulates the HTTP response for a CreateAnalysisRequest
type CreateAnalysisResponse struct { // {{{
	ID   []byte
	Dims life.Dimensions
}

// newCreateAnalysisResponse creates a CreateAnalysisResponse object for the given biologist
func newCreateAnalysisResponse(biologist *biologist.Biologist) *CreateAnalysisResponse {
	resp := new(CreateAnalysisResponse)

	resp.ID = biologist.ID
	resp.Dims = biologist.Life.Dimensions()

	return resp
} // }}}

type patternType int // {{{

const (
	// USER specifies that the pattern was passed in by the user in the request
	USER patternType = iota
	// RANDOM specifies that a random pattern should be generated
	RANDOM
) // }}}

// CreateAnalysisRequest encapsulates the needed initial data for starting a life simulation
type CreateAnalysisRequest struct { // {{{
	Dims    life.Dimensions
	pattern patternType
	Seed    []life.Location
	// life.Rules
}

func (t *CreateAnalysisRequest) String() string {
	var buf bytes.Buffer

	buf.WriteString(t.Dims.String())

	return buf.String()
}

func createAnalysis(mgr *biologist.Manager, log *log.Logger, w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		panic(err)
	}

	if err := r.Body.Close(); err != nil {
		panic(err)
	}

	var req CreateAnalysisRequest
	if err := json.Unmarshal(body, &req); err != nil {
		log.Printf("ERROR: Could not handle request: %s\n", err)
		log.Printf("REQ: %s\n", body)
		postJSON(w, 422, err)
	} else {
		// log.Printf("Received create request: %s\n", req.String())

		// Determine the pattern to use for seeding the board
		var patternFunc func(life.Dimensions, life.Location) []life.Location
		switch req.pattern {
		case USER:
			patternFunc = func(dims life.Dimensions, offset life.Location) []life.Location {
				return req.Seed
			}
		case RANDOM:
			patternFunc = func(dims life.Dimensions, offset life.Location) []life.Location {
				return life.Random(dims, offset, 35)
			}
		}

		// Create the biologist
		// log.Printf("Creating new biologist with pattern: %v\n", patternFunc(req.Dims, life.Location{X: 0, Y: 0}))
		biologist, err := biologist.New(req.Dims, patternFunc, life.ConwayTester())
		if err != nil {
			panic(err)
		}
		mgr.Add(biologist)
		// log.Println(biologist)

		// Respond the request with the ID of the biologist
		resp := newCreateAnalysisResponse(biologist)

		postJSON(w, http.StatusCreated, resp)
	}
} // }}}

/////////////////////////////////// UPDATE ANALYSIS ///////////////////////////////////

// AnalysisUpdate encapsulates the analysis of a given generation
type AnalysisUpdate struct { // {{{
	ID         []byte
	Dims       life.Dimensions
	Status     string
	Generation int
	Living     []life.Location
	// Changes    []biologist.ChangedLocation
}

func newAnalysisUpdate(biologist *biologist.Biologist, generation int) *AnalysisUpdate {
	analysis := biologist.Analysis(generation)
	if analysis == nil {
		return nil
	}

	a := new(AnalysisUpdate)

	a.ID = biologist.ID
	a.Dims = biologist.Life.Dimensions()
	a.Generation = generation

	a.Status = analysis.Status.String()

	a.Living = make([]life.Location, len(analysis.Living))
	copy(a.Living, analysis.Living)

	// a.Changes = make([]biologist.ChangedLocation, len(analysis.Changes))
	// copy(a.Changes, analysis.Changes)

	return a
} // }}}

// AnalysisUpdateRequest encapsulates required parameters to retrieve the analysis of a range of generations
type AnalysisUpdateRequest struct { // {{{
	ID                 []byte
	StartingGeneration int
	NumMaxGenerations  int
}

func (t *AnalysisUpdateRequest) String() string {
	var buf bytes.Buffer

	buf.WriteString("ID: ")
	buf.WriteString(fmt.Sprintf("%x", t.ID))
	buf.WriteString("\nStarting Generation: ")
	buf.WriteString(strconv.Itoa(t.StartingGeneration))
	buf.WriteString("\nMax: ")
	buf.WriteString(strconv.Itoa(t.NumMaxGenerations))

	return buf.String()
} // }}}

// AnalysisUpdateResponse encapsulates the requested analysis updates
type AnalysisUpdateResponse struct { // {{{
	ID      []byte
	Updates []AnalysisUpdate
}

func newAnalysisUpdateResponse(log *log.Logger, biologist *biologist.Biologist, startingGeneration int, maxGenerations int) *AnalysisUpdateResponse {
	// fmt.Printf("NewAnalysisUpdateResponse(%d, %d)\n", startingGeneration, maxGenerations)
	if biologist == nil {
		return nil
	}

	r := new(AnalysisUpdateResponse)

	r.ID = biologist.ID

	r.Updates = make([]AnalysisUpdate, 0)

	// Retrieve as many updates as are available up to the max
	endGen := startingGeneration + maxGenerations
	for i := startingGeneration; i < endGen; i++ {
		// fmt.Printf(">> Generation %d living <<\n", i)
		update := newAnalysisUpdate(biologist, i)
		// for j, change := range update.Changes {
		// fmt.Printf("  Change[%d] = %s\n", j, change.String())
		// }
		if update != nil {
			r.Updates = append(r.Updates, *update)
		}
	}
	log.Printf("Sending %d updates starting at generation %d\n", len(r.Updates), startingGeneration)

	return r
} // }}}

func getAnalysisStatus(mgr *biologist.Manager, log *log.Logger, w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		panic(err)
	}

	if err := r.Body.Close(); err != nil {
		panic(err)
	}

	var req AnalysisUpdateRequest

	// fmt.Println(string(body))
	if err := json.Unmarshal(body, &req); err != nil {
		log.Printf("ERROR: Could not handle request: %s\n", err)
		log.Printf("request: %s\n", req.String())
		postJSON(w, 422, err)
	} else {
		// log.Printf("Received poll request: %s\n", req.String())
		if resp := newAnalysisUpdateResponse(log, mgr.Biologist(req.ID), req.StartingGeneration, req.NumMaxGenerations); resp != nil {
			postJSON(w, http.StatusCreated, resp)
		}
	}
}

/////////////////////////////////// CONTROL ANALYSIS ///////////////////////////////////

// ControlOrder enumerates the available commands to control a running simulation/analysis
type ControlOrder int // {{{

const (
	// Start will begin the simulation and subsequent analysis
	Start ControlOrder = 0
	// Stop will stop the simulation (and analysis)
	Stop ControlOrder = 1
)

// ControlRequest encapsulates the HTTP request for controlling a simulation
type ControlRequest struct {
	ID    []byte
	Order ControlOrder
}

func (t *ControlRequest) String() string {
	var buf bytes.Buffer
	buf.WriteString("ID: ")
	buf.WriteString(fmt.Sprintf("%x", t.ID))
	buf.WriteString("\nOrder: ")
	switch t.Order {
	case 0:
		buf.WriteString("Start")
	case 1:
		buf.WriteString("Stop")
	}

	return buf.String()
} // }}}

func controlAnalysis(mgr *biologist.Manager, log *log.Logger, w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		panic(err)
	}

	if err := r.Body.Close(); err != nil {
		panic(err)
	}

	var req ControlRequest

	if err := json.Unmarshal(body, &req); err != nil {
		postJSON(w, 422, err)
	} else {
		log.Printf("Received control request: %s\n", req.String())

		biologist := mgr.Biologist(req.ID)

		switch req.Order {
		case Start:
			biologist.Start()
		case Stop:
			biologist.Stop()
		}
	}
}

/////////////////////////////////// OTHER ///////////////////////////////////

func postJSON(w http.ResponseWriter, httpStatus int, send interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Methods", "PUT")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")

	w.WriteHeader(httpStatus)
	if err := json.NewEncoder(w).Encode(send); err != nil {
		panic(err)
	}
}

func main() {
	logger := log.New(os.Stdout, "[biologistd] ", 0)
	portPtr := flag.Int("port", 8081, "Specify the port to use")
	flag.Parse()

	mux := http.NewServeMux()

	mgr := biologist.NewManager()

	mux.HandleFunc("/analyze",
		func(w http.ResponseWriter, r *http.Request) {
			createAnalysis(mgr, logger, w, r)
		})
	mux.HandleFunc("/poll",
		func(w http.ResponseWriter, r *http.Request) {
			getAnalysisStatus(mgr, logger, w, r)
		})
	mux.HandleFunc("/control",
		func(w http.ResponseWriter, r *http.Request) {
			controlAnalysis(mgr, logger, w, r)
		})

	http.ListenAndServe(fmt.Sprintf(":%d", *portPtr), mux)
}

// vim: set foldmethod=marker:
