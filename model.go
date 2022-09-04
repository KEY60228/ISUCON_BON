package main

import (
	"sync"
	"time"

	"github.com/isucon/isucandar/agent"
)

type User struct {
	mu sync.RWMutex

	ID          int       `json:"id"`
	AccountName string    `json:"account_name"`
	Password    string    `json:"password"`
	Authority   int       `json:"authority"`
	DeleteFlag  int       `json:"del_flg"`
	CreatedAt   time.Time `json:"created_at"`

	Agent *agent.Agent
}

func (m *User) GetAgent(o Option) (*agent.Agent, error) {
	m.mu.RLock()
	a := m.Agent
	m.mu.RUnlock()

	if a != nil {
		return a, nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	a, err := o.NewAgent(false)
	if err != nil {
		return nil, err
	}
	m.Agent = a

	return a, nil
}

type Post struct {
	ID          int       `json:"id"`
	Mime        string    `json:"mime"`
	Body        string    `json:"body"`
	ImgdataHash string    `json:"imgdata_hash"`
	UserID      int       `json:"user_id"`
	CreatedAt   time.Time `json:"created_at"`
}

type Comment struct {
	ID        int       `json:"id"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
	PostID    int       `json:"post_id"`
	UserID    int       `json:"user_id"`
}
