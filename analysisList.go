package biologist

type analysisListAddOp struct {
	analysis Analysis
	resp     chan bool
}

type analysisListGetOp struct {
	index int
	resp  chan Analysis
}

type analysisListGetAllOp struct {
	resp chan []Analysis
}

type analysisListCountOp struct {
	resp chan int
}

type analysisList struct {
	analysisListAdd    chan *analysisListAddOp
	analysisListGet    chan *analysisListGetOp
	analysisListGetAll chan *analysisListGetAllOp
	analysisListCount  chan *analysisListCountOp
}

func (t *analysisList) list() {
	var list = make([]Analysis, 0)

	for {
		select {
		case add := <-t.analysisListAdd:
			added := true
			list = append(list, add.analysis)
			add.resp <- added
		case get := <-t.analysisListGet:
			get.resp <- list[get.index]
		case getall := <-t.analysisListGetAll:
			all := make([]Analysis, 0)
			for _, val := range list {
				all = append(all, val)
			}
			getall.resp <- all
		case countOp := <-t.analysisListCount:
			countOp.resp <- len(list)
		}
	}
}

func (t *analysisList) Add(analysis Analysis) bool {
	add := &analysisListAddOp{analysis: analysis, resp: make(chan bool)}
	t.analysisListAdd <- add
	val := <-add.resp

	return val
}

func (t *analysisList) Get(idx int) Analysis {
	get := &analysisListGetOp{index: idx, resp: make(chan Analysis)}
	t.analysisListGet <- get
	val := <-get.resp

	return val
}

func (t *analysisList) GetAll() []Analysis {
	get := &analysisListGetAllOp{resp: make(chan []Analysis)}
	t.analysisListGetAll <- get
	val := <-get.resp

	return val
}

func (t *analysisList) Count() int {
	count := &analysisListCountOp{resp: make(chan int)}
	t.analysisListCount <- count
	val := <-count.resp

	return val
}

// func (t *analysisList) Clone() *analysisList {
// 	shadow := newAnalysisList()
//
// 	// FIXME: copy the analysisList using t.GetAll()
//
// 	return shadow
// }

func newAnalysisList() *analysisList {
	t := new(analysisList)

	t.analysisListAdd = make(chan *analysisListAddOp)
	t.analysisListGet = make(chan *analysisListGetOp)
	t.analysisListGetAll = make(chan *analysisListGetAllOp)
	t.analysisListCount = make(chan *analysisListCountOp)

	go t.list()

	return t
}

// vim: set foldmethod=marker:
