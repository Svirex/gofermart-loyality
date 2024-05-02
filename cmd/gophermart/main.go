package main

import (
	"encoding/json"
	"fmt"
)

type J struct {
	Value float64 `json:",omitempty"`
}

func main() {
	j := J{
		Value: float64(0.00),
	}
	v, _ := json.Marshal(&j)
	fmt.Println(string(v))
}

// 8 3 3 4
// 7 3 6 4
