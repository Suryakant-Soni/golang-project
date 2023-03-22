package worklist

type Entry struct {
	Path string
}

type Worklist struct {
	// this channel stores the work or jobs to be done
	jobs chan Entry
}

// add a new job which has come to the channel
func (w *Worklist) Add(work Entry) {
	w.jobs <- work
}
// find out the next job to be done
func (w *Worklist) Next() Entry {
	j := <-w.jobs
	return j
}
// constructor for the worklist struct basically
func New(bufsize int) Worklist {
	return Worklist{make(chan Entry, bufsize)}
}
// constructor for a new job or entry
func NewJob(path string) Entry {
	return Entry{path}
}
// sending empty signal to Jobs channel to shut the  routine
func (w *Worklist) Finalize(numWorkers int) {
	for i := 0; i < numWorkers; i++ {
		w.Add(Entry{""})
	}
}
