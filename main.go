package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
)

type record struct {
	Q string
	A string
}

type quiz struct {
	Records []record
	Score   int
}

var currentDirectory, _ = os.Getwd()
var defaultQuizFile = strings.Join([]string{currentDirectory, "resources", "problems.csv"}, string(os.PathSeparator))
var quizFilePath string
var timeout time.Duration

func init() {
	flag.DurationVar(&timeout, "timeout", time.Duration(20) * time.Second, "Quiz duration in seconds")
	flag.StringVar(&quizFilePath, "path", defaultQuizFile, "Path to a csv quiz file")
	flag.Parse()
}

func initQuiz() (*quiz, error) {
	file, err := os.Open(quizFilePath)
	if err != nil {
		fmt.Printf("failed reading the quiz file %s", quizFilePath)
		return nil, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Printf("failed closing the quiz file %s", quizFilePath)
			return
		}
	}(file)

	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		fmt.Printf("failed parsing quiz file %s", quizFilePath)
		return nil, err
	}

	var records []record
	for _, row := range rows {
		records = append(records, record{Q: row[0], A: row[1]})
	}
	res :=  &quiz{Records: shuffle(records), Score: 0}
	fmt.Printf("Math quiz. Complete %d questions in %d seconds\n", len(records), int(timeout.Seconds()))
	return res, nil
}

func shuffle(records []record) []record {
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(records), func(i, j int) {
		records[i], records[j] = records[j], records[i]
	})
	return records
}

func (qz *quiz) askQ(q record, s *bufio.Scanner) {
	fmt.Printf("%s = ", q.Q)
	if s.Scan() {
		input := strings.TrimSpace(s.Text())
		if input == q.A {
			qz.Score++
		}
	}
}

func printResults(qz *quiz, timesUp bool) {
	score := qz.Score*100/len(qz.Records)
	if timesUp {
		fmt.Print("\nTimes up, ")
	}
	fmt.Printf("%d/%d of the questions were correctly answered. Your score is %d\n", qz.Score, len(qz.Records), score)
}

func main() {
	fmt.Println(os.Getwd())
	qz, err := initQuiz()
	if err != nil {
		fmt.Printf("failed initializing the quiz. cause: %s", err)
		return
	}
	doneChannel := make(chan bool)
	timeoutChannel := time.After(timeout)
	s := bufio.NewScanner(os.Stdin)
	go func() {
		for _, rec := range qz.Records {
			qz.askQ(rec, s)
		}
		doneChannel <- true
	}()
	select {
	case <-doneChannel:
	case <-timeoutChannel:
		printResults(qz, true)
		return
	}
	printResults(qz, false)
}