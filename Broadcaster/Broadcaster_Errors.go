package Broadcaster

import (
	"fmt"
)

func CheckforErrors(err error) {
	if err != nil {
		fmt.Println(err)
	}
}
