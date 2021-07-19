/**
 * Copyright 2020 gd Author. All rights reserved.
 * Author: Chuck1024
 */
package gr

import "sync"

/**
Notice:
1. it use lock to make sure taskList is correct. performance is bad
2. still able to add task after WaitStopAll, so you should make sure no extra task is running after called WaitStopAll
*/
type StopAbleTask interface {
	Close()
}

type StopAbleTaskGroup struct {
	taskList []StopAbleTask
	lock     sync.Mutex
}

func NewStopAbleTaskGroup() *StopAbleTaskGroup {
	return &StopAbleTaskGroup{
		taskList: make([]StopAbleTask, 0),
	}
}

func (s *StopAbleTaskGroup) AddTask(task StopAbleTask) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.taskList == nil {
		s.taskList = make([]StopAbleTask, 0)
	}
	s.taskList = append(s.taskList, task)
}

func (s *StopAbleTaskGroup) Start() error {
	if s.taskList == nil {
		s.taskList = make([]StopAbleTask, 0)
	}
	return nil
}

func (s *StopAbleTaskGroup) Close() {
	s.WaitStopAll()
}

func (s *StopAbleTaskGroup) WaitStopAll() {
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
	s.taskList = make([]StopAbleTask, 0)
}
