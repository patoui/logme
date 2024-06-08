package listener

import (
	"fmt"
)

type LogCreated struct {
}

func (e LogCreated) Handle(data string) {
	fmt.Printf("HANDLE LOG CREATED:\n%v\n", data)
}
