package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

type Model interface {
	GetID() int
	GetCreatedAt() time.Time
}

type Set[T Model] struct {
	mu   sync.RWMutex
	list []T
	dict map[int]T
}

func (s *Set[T]) At(index int) T {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.list == nil {
		return *new(T)
	}

	return s.list[index]
}

func (s *Set[T]) Get(id int) (T, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.dict == nil {
		return *new(T), false
	}

	model, ok := s.dict[id]
	return model, ok
}

func (s *Set[T]) Add(model T) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := model.GetID()
	if id == 0 {
		return false
	}

	if len(s.list) == 0 {
		s.list = []T{model}
	} else {
		pos := 0
		for i := 0; i < len(s.list)-1; i++ {
			m := s.list[i]
			pos = i

			if m.GetCreatedAt().Equal(model.GetCreatedAt()) {
				if m.GetID() > model.GetID() {
					break
				}
			}
		}

		s.list = append(s.list[:pos+1], s.list[pos:]...)
		s.list[pos] = model
	}

	if s.dict == nil {
		s.dict = make(map[int]T, 0)
	}
	s.dict[id] = model

	return true
}

func (s *Set[T]) LoadJSON(jsonFile string) error {
	file, err := os.Open(jsonFile)
	if err != nil {
		return err
	}
	defer file.Close()

	models := []T{}

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&models); err != nil {
		return err
	}

	for _, model := range models {
		if !s.Add(model) {
			return fmt.Errorf("unexpected error on dump loading: %v", model)
		}
	}

	return nil
}
