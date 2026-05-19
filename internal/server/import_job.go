package server

import "sync"

type importJob struct {
	mu      sync.Mutex
	running bool
	stage   string
	message string
	rows    int64
	err     string
}

func (j *importJob) snapshot() (running bool, stage, message, err string, rows int64) {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.running, j.stage, j.message, j.err, j.rows
}

func (j *importJob) start(stage, message string) bool {
	j.mu.Lock()
	defer j.mu.Unlock()
	if j.running {
		return false
	}
	j.running = true
	j.stage = stage
	j.message = message
	j.rows = 0
	j.err = ""
	return true
}

func (j *importJob) set(stage, message string) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.stage = stage
	j.message = message
}

func (j *importJob) finish(rows int64, err error) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.running = false
	j.rows = rows
	if err != nil {
		j.stage = "error"
		j.err = err.Error()
		j.message = "Import failed"
		return
	}
	j.stage = "done"
	j.err = ""
	j.message = "Import complete"
}
