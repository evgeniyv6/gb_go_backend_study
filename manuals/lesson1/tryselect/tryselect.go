package tryselect

import (
	"fmt"
	"log"
	"time"
)

type Data struct {
	num   int
	descr string
}

func NewData() *Data {
	return &Data{3, "this is ten"}
}

func expansiveComputations(data *Data, answer chan int, done chan bool) {
	var res int
	allDone := false
	for !allDone {
		for i := 0; i < data.num; i++ {
			time.Sleep(1 * time.Second)
			res += i
		}
		allDone = true
		answer <- res
	}
	done <- true
}

func CustExec() {
	const allDone = 2
	doneCount := 0

	answ1 := make(chan int)
	answ2 := make(chan int)

	defer func() {
		close(answ1)
		close(answ2)
	}()

	done := make(chan bool)

	defer func() { close(done) }()

	go expansiveComputations(NewData(), answ1, done)
	go expansiveComputations(NewData(), answ2, done)

	for doneCount != allDone {
		var which, res int

		select {
		case res = <-answ1:
			which = 1
		case res = <-answ2:
			which = 2
		case <-done:
			doneCount++
		}

		if which != 0 {
			log.Printf("%c ->%d", which, res)
		}
	}
	fmt.Println()
}
