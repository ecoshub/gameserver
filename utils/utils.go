package utils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func InterruptHandle() {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// signal catch routine
	go func() {
		<-signalChannel
		fmt.Println("*******************")
		time.Sleep(time.Microsecond * 100)
		os.Exit(0)
	}()
}

func PrintStruct(i interface{}) {
	enc, _ := json.MarshalIndent(i, "", "  ")
	fmt.Println(string(enc))
}

func SprintStruct(i interface{}) string {
	enc, _ := json.MarshalIndent(i, "", "  ")
	return string(enc)
}

func RandomSleepMillisecond(min, max int) {
	t := rand.Intn(max-min) + min
	fmt.Println(t)
	time.Sleep(time.Millisecond * time.Duration(t))
}

func Halt() {
	for {
		time.Sleep(time.Hour)
	}
}

func ReadNBytes(b *bufio.Reader, n int) ([]byte, error) {
	buffer := make([]byte, 0, n)
	for i := 0; i < n; i++ {
		singleByte, err := b.ReadByte()
		if err != nil {
			return []byte{}, err
		}
		buffer = append(buffer, singleByte)
	}
	return buffer, nil
}
