package main

import (
	"fmt"
	"os"
)

func CheckforErrors(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
