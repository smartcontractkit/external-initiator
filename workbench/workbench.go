package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("started")
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("recovered", r)
			}
		}()
		panic("err")
		fmt.Println("continuing?")
	}()
	time.Sleep(1 * time.Second)
	fmt.Println("done")
}
