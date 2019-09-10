package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
)

type version map[string]string

func compare(v, w version) int {
	comp := func(s1, s2 string) int64 {
		a, _ := strconv.ParseInt(s1, 10, 32)
		b, _ := strconv.ParseInt(s2, 10, 32)
		return a - b
	}

	if c := comp(v["major"], w["major"]); c < 0 {
		return -1
	} else if c > 0 {
		return 1
	}

	log("equal major")

	if c := comp(v["minor"], w["minor"]); c < 0 {
		return -1
	} else if c > 0 {
		return 1
	}

	log("equal minor")

	if c := comp(v["patch"], w["patch"]); c < 0 {
		return -1
	} else if c > 0 {
		return 1
	}

	log("equal patch")

	if v["prelease"] == "" && w["prerelease"] != "" {
		return 1
	}

	if v["prelease"] != "" && w["prerelease"] == "" {
		return -1
	}

	log("both pre-release")
	return 0
}

func (v version) String() string {
	str := fmt.Sprintf("v%s.%s.%s", v["major"], v["minor"], v["patch"])
	if v["prerelease"] != "" {
		str += "-" + v["prerelease"]
	}
	if v["buildmetadata"] != "" {
		str += "+" + v["buildmetadata"]
	}
	return str
}

func parse(str string) version {
	// Regex for semver: https://semver.org/
	// Example: https://regex101.com/r/Ly7O1x/3/
	r := regexp.MustCompile("^(?P<major>0|[1-9]\\d*)\\.(?P<minor>0|[1-9]\\d*)\\.(?P<patch>0|[1-9]\\d*)(?:-(?P<prerelease>(?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\\.(?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\\.[0-9a-zA-Z-]+)*))?$")
	v := version{}
	groups := r.SubexpNames()
	matches := r.FindStringSubmatch(str)
	for i := 1; i < len(matches); i++ {
		v[groups[i]] = matches[i]
	}
	return v
}

func compareStr(a, b string) int {
	return compare(parse(a), parse(b))
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Wrong number of arguments, should be two!")
		os.Exit(1)
	}
	v := parse(os.Args[1])
	w := parse(os.Args[2])
	c := compare(v, w)
	switch c {
	case -1:
		fmt.Printf("'%s' precedes '%s'\n", v.String(), w.String())
	case 0:
		fmt.Printf("'%s' is equivalent to '%s'\n", v.String(), w.String())
	case 1:
		fmt.Printf("'%s' comes after '%s'\n", v.String(), w.String())
	}
}

func log(msg string) {
	fmt.Println("log:", msg)
}
