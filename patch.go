package scimpatch

import (
	"fmt"
	"strings"
)

type Operation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}

func NewPatch(operations []Operation) func(src interface{}) (interface{}, error) {
	// Some SCIM clients such as AzureAD sends `op` as PascalCase so here changes them to lowercase
	var _operations []Operation
	copy(_operations, operations)
	for _, op := range _operations {
		op.Op = strings.ToLower(op.Op)
	}

	return func(src interface{}) (interface{}, error) {
		for _, op := range _operations {
			switch op.Op {
			case "add":
				fmt.Println("add")
			case "remove":
				fmt.Println("remove")
			case "replace":
				fmt.Println("replace")
			}
		}

		//
		// TODO
		//
		return fmt.Sprintf("%s", src), nil
	}
}
