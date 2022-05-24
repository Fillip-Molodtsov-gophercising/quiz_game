package main

import (
	"flag"
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"math/rand"
	"strings"
	"time"

	"os"
	"path/filepath"
)

type QAItem struct {
	Q string
	A string
}

type QuizData struct {
	QuizList []QAItem `yaml:"quizList"`
}

func main() {
	path, timer := setFlags()
	qd := QuizData{}
	extractFromYml(path, &qd)
	qd = adjustGivenAnswers(qd)
	startGame(qd, *timer)
}

func startGame(qd QuizData, timeCount int) {
	var (
		answer   string
		count    int
		timeLeft = timeCount
		winText  = "The result is: %d/%d\n"
		loseTest = fmt.Sprintf("You run out of time. %s", winText)
		done     = make(chan struct{})
		lost     = make(chan struct{})
	)
	fmt.Print("To start the timeCount press any key")
	fmt.Scanln(&answer)
	ticker := time.NewTicker(1 * time.Second)
	go promptingQuestions(qd, done, &answer, &count, &timeLeft)
	go func() {
		for timeLeft > 0 {
		}
		close(lost)
	}()

Loop:
	for {
		select {
		case <-done:
			fmt.Printf(winText, count, len(qd.QuizList))
			ticker.Stop()
			break Loop
		case <-lost:
			fmt.Printf(loseTest, count, len(qd.QuizList))
			ticker.Stop()
			break Loop
		case <-ticker.C:
			timeLeft--
		}
	}
}

func promptingQuestions(qd QuizData, done chan struct{}, answer *string, count *int, timeLeft *int) {
	for _, qa := range qd.QuizList {
		fmt.Println(qa.Q)
		fmt.Scanln(answer)
		strings.TrimSpace(*answer)
		if *answer != qa.A {
			continue
		}
		*count++
		*timeLeft = *timeLeft + 1
	}
	close(done)
}

func adjustGivenAnswers(qd QuizData) QuizData {
	for i, qa := range qd.QuizList {
		qd.QuizList[i].A = strings.ToLower(strings.TrimSpace(qa.A))
	}
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(qd.QuizList),
		func(i, j int) { qd.QuizList[i], qd.QuizList[j] = qd.QuizList[j], qd.QuizList[i] })
	return qd
}

func extractFromYml(path *string, out *QuizData) {
	fc, err := os.ReadFile(*path)
	if err != nil {
		log.Fatalf("Some error occured while trying to read this file: %s\n%v", *path, err)
	}

	err = yaml.Unmarshal(fc, out)
	if err != nil {
		log.Fatalf("Cannot unmarshal the content of yaml: %v", err)
	}
}

func setFlags() (*string, *int) {
	defer flag.Parse()
	path := flag.String("file", defaultYml(), "The file where Q&A are stated")
	timer := flag.Int("timer", 2, "Set the timeout for the game")
	return path, timer
}

func defaultYml() string {
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Cannot find working directory: %v", err)
	}
	return filepath.Join(pwd, "example.yml")
}
