package main

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"os"
	"time"
)

type Score struct {
	Id       int
	AnswerId int `db:"answerId"`
	Score    int
	MarkerId int `db:"markerId"`
}

type User struct {
	Id       int
	Name     string
	Password string
	Email    string
	IsAdmin  bool      `db:"isAdmin"`
	CreateAt time.Time `db:"createAt"`
	UpdateAt time.Time `db:"updateAt"`
}

type Question struct {
	Id       int
	Title    string
	Content  string
	Hidden   bool
	CreateAt time.Time `db:"createAt"`
	UpdateAt time.Time `db:"updateAt"`
}

type Answer struct {
	Id         int
	Content    string
	UserId     int       `db:"userId"`
	QuestionId int       `db:"questionId"`
	CreateAt   time.Time `db:"createAt"`
	UpdateAt   time.Time `db:"updateAt"`
}

type Status struct {
	Count   int64
	Done    int64
	Percent float64
}

var db *sqlx.DB
var schema = `
CREATE TABLE IF NOT EXISTS score (
	id SERIAL primary key,
	"answerId" integer NOT NULL,
	score integer NOT NULL,
	"markerId" integer NOT NULL
);

`

func connectDb() {
	var err error
	db, err = sqlx.Connect("postgres", "host=localhost user=root password=root dbname=pg port=5432 sslmode=disable TimeZone=Asia/Shanghai")
	if err != nil {
		fmt.Println("database connect error!")
		os.Exit(-1)
	}
	db.MustExec(schema)
}

func queryAllQuestions() *[]Question {
	var result []Question
	err := db.Select(&result, "SELECT * FROM question")
	if err != nil {
		fmt.Println(err)
	}
	return &result
}
func queryQuestionById(questionId int) *Question {
	var result Question
	err := db.Get(&result, `SELECT * FROM question WHERE id = $1`, questionId)
	if err != nil {
		fmt.Println(err)
	}
	return &result
}

func queryAllAnswer(questionId int) *[]Answer {
	var result []Answer
	err := db.Select(&result, `SELECT * FROM answer WHERE "questionId" = $1 ORDER BY id`, questionId)
	if err != nil {
		fmt.Println(err)
	}
	return &result
}

func queryScore(answerId int) *Score {
	var result Score
	err := db.Get(&result, `SELECT * FROM score WHERE "answerId" = $1`, answerId)
	if err != nil {
		fmt.Println(err)
	}
	return &result
}

func createScore(score *Score) {
	tx := db.MustBegin()
	var s Score
	tx.Get(&s, `SELECT * FROM score WHERE "answerId" = $1`, score.AnswerId)
	if s.Id == 0 {
		tx.MustExec(`INSERT INTO score ("answerId",score,"markerId") VALUES ($1,$2,$3)`, score.AnswerId, score.Score, score.MarkerId)
	} else {
		tx.MustExec(`UPDATE score SET score = $1,"markerId" = $2 WHERE id = $3`, score.Score, score.MarkerId, s.Id)
	}
	tx.Commit()
}

func queryUser(email string) *User {
	var result User
	err := db.Get(&result, `SELECT * FROM "user" WHERE email = $1`, email)
	if err != nil {
		fmt.Println(err)
	}
	return &result
}

func queryUserById(id int) *User {
	var result User
	err := db.Get(&result, `SELECT * FROM "user" WHERE id = $1`, id)
	if err != nil {
		fmt.Println(err)
	}
	return &result
}

func queryJudgeProgress(questionId int) Status {
	var count int64
	var done int64
	err := db.Get(&count, `SELECT count(1) FROM answer WHERE "questionId" = $1`, questionId)
	if err != nil {
		fmt.Println(err)
	}
	err = db.Get(&done, `SELECT count(1) FROM answer RIGHT JOIN score ON score."answerId" = answer.id WHERE "questionId" = $1`, questionId)
	if err != nil {
		fmt.Println(err)
	}
	return Status{count, done, float64(done) / float64(count) * 100}
}

func queryAnswers(userId int) *[]Answer {
	var result []Answer
	err := db.Select(&result, `SELECT * FROM answer WHERE "userId" = $1 ORDER BY "questionId"`, userId)
	if err != nil {
		fmt.Println(err)
	}
	return &result
}
