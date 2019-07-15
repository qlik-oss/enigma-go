package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	major int = iota
	minor
	patch
)

// parse parses a version and bumps it according to the specified flag.
// No support for maintaining meta data (+) or info such as -beta/-rc,
// the scope of this program is only to bump the version according to the
// supplied flag.
func parse(arg string) ([3]int, error) {
	// Assuming the information after "-" is just an unabbreviated sha from
	// git describe, will be ignored
	split := strings.Split(arg, "-")
	split = strings.Split(split[0], ".")
	ver := [3]int{}
	if len(split) != 3 {
		return ver, fmt.Errorf("a semantic version must be of the form 1.2.3")
	}
	// If it begins with 'v' remove it so we can parse it as an int.
	if split[0][0] == 'v' {
		split[0] = split[0][1:]
	}
	for i, s := range split {
		v, err := strconv.Atoi(s)
		if err != nil {
			return ver, err
		}
		ver[i] = v
	}
	return ver, nil
}

// bump bumps the version according to the flag and sets all following numbers
// in the version to zero.
func bump(ver [3]int, flag int) string {
	switch flag {
	case major:
		ver[0]++
		ver[1] = 0
		ver[2] = 0
	case minor:
		ver[1]++
		ver[2] = 0
	case patch:
		ver[2]++
	}
	version := fmt.Sprintf("%d.%d.%d", ver[0], ver[1], ver[2])
	return version
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Invalid number of arguments, usage: go run bump.go 1.2.3 -M|m|p")
		fmt.Println("\n  -M major")
		fmt.Println("  -m minor")
		fmt.Println("  -p patch")
		fmt.Println("\nExample: 'go run bump.go 1.2.3 -p' => '1.2.4'")
		return
	}
	var flag int
	switch arg := os.Args[2]; arg {
	case "-M":
		flag = major
	case "-m":
		flag = minor
	case "-p":
		flag = patch
	default:
		fmt.Printf("Invalid second argument '%s', must be one of '-M', '-m' or '-p'\n", arg)
		return
	}
	ver, err := parse(os.Args[1])
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	version := bump(ver, flag)
	fmt.Println(version)
}
