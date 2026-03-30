package prompt

import (
	"errors"
	"fmt"
	"os"

	"charm.land/huh/v2"
)

func checkAbort(err error) {
	if err == nil {
		return
	}
	if errors.Is(err, huh.ErrUserAborted) {
		fmt.Fprintln(os.Stderr, "\n  Aborted.")
		os.Exit(130)
	}
	fmt.Fprintf(os.Stderr, "\n  Error: %v\n", err)
	os.Exit(1)
}

func Confirm(message string, defaultYes bool) bool {
	result := defaultYes
	err := huh.NewConfirm().
		Title(message).
		Affirmative("Yes").
		Negative("No").
		Value(&result).
		Run()
	checkAbort(err)
	return result
}

func Select(message string, options []string) int {
	opts := make([]huh.Option[int], len(options))
	for i, o := range options {
		opts[i] = huh.NewOption(o, i)
	}

	var result int
	err := huh.NewSelect[int]().
		Title(message).
		Options(opts...).
		Height(len(options) + 1).
		Value(&result).
		Run()
	checkAbort(err)
	return result
}

// SelectMultiple presents options with an "All" choice at the top.
// Returns indices from the original options slice.
func SelectMultiple(message string, options []string) []int {
	allIndices := make([]int, len(options))
	for i := range options {
		allIndices[i] = i
	}

	opts := make([]huh.Option[int], 0, len(options)+1)
	opts = append(opts, huh.NewOption("All", -1))
	for i, o := range options {
		opts = append(opts, huh.NewOption(o, i))
	}

	var result int
	err := huh.NewSelect[int]().
		Title(message).
		Options(opts...).
		Height(len(opts) + 1).
		Value(&result).
		Run()
	checkAbort(err)

	if result == -1 {
		return allIndices
	}
	return []int{result}
}
