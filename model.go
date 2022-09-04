package main

import "time"

type User struct {
	ID          int       `json:"id"`
	AccountName string    `json:"account_name"`
	Password    string    `json:"password"`
	Authority   int       `json:"authority"`
	DeleteFlag  int       `json:"del_flg"`
	CreatedAt   time.Time `json:"created_at"`
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
