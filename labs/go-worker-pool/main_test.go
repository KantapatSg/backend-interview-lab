package main

import (
	"context"
	"sort"
	"testing"
)

func TestRunPoolProcessesEveryInput(t *testing.T) {
	got, err := runPool(context.Background(), []int{1, 2, 3, 4}, 2)
	if err != nil {
		t.Fatal(err)
	}
	sort.Slice(got, func(i, j int) bool { return got[i].input < got[j].input })
	for i, item := range got {
		input := i + 1
		if item.input != input || item.square != input*input {
			t.Fatalf("result[%d] = %+v", i, item)
		}
	}
}
