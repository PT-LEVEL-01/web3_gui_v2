package task

import (
	"sync"
	"time"
)

type TaskManager struct {
	taskMap map[string]*Task
}

var TaskM *TaskManager

var Once_Task sync.Once

func init() {
	Once_Task.Do(func() {
		TaskM = &TaskManager{make(map[string]*Task)}
	})
}

func (tm *TaskManager) Exists(k string) bool {
	_, ok := tm.taskMap[k]
	return ok
}

func (tm *TaskManager) AddAndDo(k string, t *Task) bool {
	if tm.Exists(k) {
		return false
	}

	tm.taskMap[k] = t

	go t.Exec()

	return true
}

func (tm *TaskManager) Add(k string, t *Task) bool {
	if tm.Exists(k) {
		return false
	}

	tm.taskMap[k] = t
	return true
}

func (tm *TaskManager) Do(k string) bool {
	if !tm.Exists(k) {
		return false
	}

	v := tm.taskMap[k]
	if v.Status() {
		return false
	}

	go v.Exec()

	return true
}

func (tm *TaskManager) Stop(k string) bool {
	if !tm.Exists(k) {
		return false
	}

	v := tm.taskMap[k]
	if !v.Status() {
		return false
	}

	v.Stop()

	return true
}

type Task struct {
	name string
	it   time.Duration
	s    bool
	m    func()
	done chan struct{}
}

func (t *Task) Status() bool {
	return t.s
}

func (t *Task) Exec() {
	t.s = true
	ticker := time.NewTicker(t.it)
	for range ticker.C {
		select {
		case <-t.done:
			t.s = false
			ticker.Stop()
			return
		default:
		}
		t.m()
	}
}

func (t *Task) Stop() {
	select {
	case t.done <- struct{}{}:
	default:
	}
}

func CreateTask(n string, f func(), t time.Duration) *Task {
	return &Task{
		name: n,
		it:   t,
		m:    f,
		done: make(chan struct{}, 1),
	}
}
