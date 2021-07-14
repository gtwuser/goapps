package main

import (
	"fmt"
	"sort"
)

func main() {
	res1 := sort.SearchStrings([]string{"apple", "banana",
		"kiwi", "orange"}, "apil")

	res2 := sort.SearchStrings([]string{"Cat", "Cow",
		"Dog", "Parrot"}, "Cat")

	// Displaying the results
	fmt.Println("Result 1: ", res1)
	fmt.Println("Result 2: ", res2)
}
