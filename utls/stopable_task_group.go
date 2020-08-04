package utls

import "sync"

/**
Notice:
1. it use lock to make sure taskList is correct. performance is bad
2. still able to add task after WaitStopAll, so you should make sure no extra task is running after called WaitStopAll
*/
type StopableTask interface {
	Close()
}

type StopableTaskGroup struct {
	taskList []StopableTask
	lock     sync.Mutex
}

func NewStopableTaskGroup() *StopableTaskGroup {
	return &StopableTaskGroup{
		taskList: make([]StopableTask, 0),
	}
}

func (s *StopableTaskGroup) AddTask(task StopableTask) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.taskList == nil {
		s.taskList = make([]StopableTask, 0)
	}
	s.taskList = append(s.taskList, task)
}

func (s *StopableTaskGroup) Start() error {
	if s.taskList == nil {
		s.taskList = make([]StopableTask, 0)
	}
	return nil
}

func (s *StopableTaskGroup) Close() {
	s.WaitStopAll()
}

func (s *StopableTaskGroup) WaitStopAll() {
	s.lock.Lock()
	defer s.lock.Unlock()

	wg := sync.WaitGroup{}
	if s.taskList != nil && len(s.taskList) > 0 {
		count := len(s.taskList)
		for i := 0; i < count; i++ {
			task := s.taskList[count-1-i]
			wg.Add(1)
			go func() {
				defer wg.Done()
				task.Close()
			}()
		}
	}
	wg.Wait()
	s.taskList = make([]StopableTask, 0)
}
