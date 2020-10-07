package main

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
)

var r = gin.Default()
var session = make(map[int]int)

func main() {
	connectDb()

	r.HTMLRender = render()
	r.GET("/", index)
	r.GET("/answers/:questionId", index)
	r.GET("/answers/:questionId/:num", answers)
	r.POST("/answers/:questionId/:num", mark)
	r.GET("/login", loginPage)
	r.POST("/login", login)
	r.Run(":10000")
}

func render() multitemplate.Renderer {
	r := multitemplate.NewRenderer()
	r.AddFromFiles("index", "templates/layout.html", "templates/index.html")
	r.AddFromFiles("login", "templates/layout.html", "templates/login.html")
	r.AddFromFiles("answer", "templates/layout.html", "templates/answer.html")
	return r
}

func checkIsLogin(c *gin.Context)bool{
	cookie,err := c.Cookie("token")
	if err != nil{
		c.Redirect(http.StatusFound,"/login")
		return false
	}
	id,_ := strconv.Atoi(cookie)
	_,exist := session[id]
	if !exist{
		c.Redirect(http.StatusFound,"/login")
		return false
	}
	return true
}

func index(c *gin.Context){
	if !checkIsLogin(c){
		return
	}
	questionId, _ := strconv.Atoi(c.Param("questionId"))

	allQuestions := *queryAllQuestions()
	progress := make(map[int]Status,0)
	for _,item := range allQuestions{
		progress[item.Id] = queryJudgeProgress(item.Id)
	}

	if questionId != 0 {
		allAnswers := *queryAllAnswer(questionId)
		c.HTML(http.StatusOK, "index" , gin.H{
			"thisQuestion" : queryQuestionById(questionId),
			"thisProgress" : queryJudgeProgress(questionId),
			"allAnswers" :      allAnswers,
		})
	}else{
		c.HTML(http.StatusOK, "index" , gin.H{
			"progress" : progress,
		})
	}
}

func answers(c *gin.Context){
	if !checkIsLogin(c){
		return
	}
	questionId, _ := strconv.Atoi(c.Param("questionId"))
	num, _ := strconv.Atoi(c.Param("num"))
	answers := *queryAllAnswer(questionId)
	score := queryScore(answers[num].Id)
	marker := queryUserById(score.MarkerId)

	nextId := num
	if len(answers) - 1 > num{
		nextId++
	}

	c.HTML(http.StatusOK, "answer" , gin.H{
		"thisQuestion" : queryQuestionById(questionId),
		"answer" :       answers[num],
		"score" :        score,
		"marker" :       marker,
		"nextId" :       nextId,
		"progress" :     queryJudgeProgress(questionId),
	})
}

func mark(c *gin.Context){
	if !checkIsLogin(c){
		return
	}
	questionId, _ := strconv.Atoi(c.Param("questionId"))
	num, _ := strconv.Atoi(c.Param("num"))
	answers := *queryAllAnswer(questionId)

	var score Score
	var err error
	var success = true
	score.Score, err = strconv.Atoi(strings.TrimSpace(c.PostForm("score")))
	if err != nil {
		success = false
	} else {
		user,_ := c.Cookie("token")
		userId,_ := strconv.Atoi(user)
		score.AnswerId = answers[num].Id
		score.MarkerId = session[userId]
		createScore(&score)
	}

	if success && len(answers) - 1 > num{
		num++
	}
	c.Redirect(http.StatusFound, "/answers/" + c.Param("questionId") + "/" + strconv.Itoa(num))
}

func loginPage(c *gin.Context){
	c.HTML(http.StatusOK, "login", gin.H{})
}

func login(c *gin.Context){
	email := c.PostForm("email")
	password := c.PostForm("password")
	hash := md5.New()
	hash.Write([]byte(password))
	password = hex.EncodeToString(hash.Sum(nil))

	user := queryUser(email)
	if user.Id == 0 || !user.IsAdmin{
		c.HTML(http.StatusOK, "login", gin.H{
			"error" : "不是管理员账号",
		})
	}else if user.Password != password{
		c.HTML(http.StatusOK, "login", gin.H{
			"error" : "密码错误",
		})
	}else {
		newSession := rand.Int()
		session[newSession] = user.Id
		c.SetCookie("token",strconv.Itoa(newSession),0,"/","",false,true)
		c.Redirect(http.StatusFound,"/")
	}
}
