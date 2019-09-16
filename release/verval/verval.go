package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
)

type version map[string]string

func compare(v, w version) int {
	if c := comp(v["major"], w["major"]); c < 0 {
		return -1
	} else if c > 0 {
		return 1
	}

	if c := comp(v["minor"], w["minor"]); c < 0 {
		return -1
	} else if c > 0 {
		return 1
	}

	if c := comp(v["patch"], w["patch"]); c < 0 {
		return -1
	} else if c > 0 {
		return 1
	}

	if v["prerelease"] == "" && w["prerelease"] != "" {
		return 1
	}

	if v["prerelease"] != "" && w["prerelease"] == "" {
		return -1
	}

	return 0
}

func comp(s1, s2 string) int64 {
	a, _ := strconv.ParseInt(s1, 10, 32)
	b, _ := strconv.ParseInt(s2, 10, 32)
	return a - b
}

func (v version) String() string {
	str := fmt.Sprintf("%s.%s.%s", v["major"], v["minor"], v["patch"])
	if v["prerelease"] != "" {
		str += "-" + v["prerelease"]
	}
	if v["buildmetadata"] != "" {
		str += "+" + v["buildmetadata"]
	}
	return str
}

func (v version) bump(field string) error {
	val, ok := v[field]
	if !ok {
		return fmt.Errorf("'%s' is an invalid field", field)
	}
	nbr, err := strconv.ParseInt(val, 10, 32)
	if err != nil {
		return fmt.Errorf("cannot bump field '%s'", field)
	}
	nbr++
	v[field] = fmt.Sprint(nbr)
	switch field {
	case "major":
		v["minor"] = "0"
		v["patch"] = "0"
	case "minor":
		v["patch"] = "0"
	}
	v["prerelease"] = ""
	v["buildmetadata"] = ""
	return nil
}

func (v version) valid() bool {
	switch {
	case v["major"] == "":
		return false
	case v["minor"] == "":
		return false
	case v["patch"] == "":
		return false
	}
	return true
}

func parse(str string) (version, error) {
	if len(str) > 0 && str[0] == 'v' {
		str = str[1:]
	}
	// Regex for semver: https://semver.org/
	// Example: https://regex101.com/r/Ly7O1x/3/
	r := regexp.MustCompile("^(?P<major>0|[1-9]\\d*)\\.(?P<minor>0|[1-9]\\d*)\\.(?P<patch>0|[1-9]\\d*)(?:-(?P<prerelease>(?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\\.(?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\\.[0-9a-zA-Z-]+)*))?$")
	v := version{}
	groups := r.SubexpNames()
	matches := r.FindStringSubmatch(str)
	for i := 1; i < len(matches); i++ {
		v[groups[i]] = matches[i]
	}
	if !v.valid() {
		return v, fmt.Errorf("'%s' is not a valid version", str)
	}
	return v, nil
}

var use = `verval can do two things, either bump a version or compare two versions.

Bumping:

  go run verval.go bump <field> <version>

where <field> can be one of 'major', 'minor' or 'patch' depending on
what you want to bump. Lastly, <version> is the semantic version which
you want to bump. Any meta or prerelease information will be removed
and is left to the user to handle.

Comparing:

  go run verval.go <version1> <version2>

where the two versions must be semantic versions. This will return
one of '1', '0' or '-1' meaning version1 is greater, equal or less
than version2 respectively. Prerelease tags are obeyed to the extent
that '1.0.0' succeeds '1.0.0-beta' but the value of the prerelease
field is not evaluated itself. ('1.0.0-beta' == '1.0.0-alpha')
`

func main() {
	if l := len(os.Args); l < 3 || l > 4 {
		exit("Expected 3 or 4 arguments.", use)
	}
	if os.Args[1] == "bump" {
		if len(os.Args) != 4 {
			exit("Bumping requires exactly 4 arguments.", use)
		}
		mode := os.Args[2]
		v, err := parse(os.Args[3])
		if err != nil {
			exit(err.Error())
		}
		v.bump(mode)
		fmt.Print(v.String())
	} else {
		v, err := parse(os.Args[1])
		if err != nil {
			exit(err.Error())
		}
		w, err := parse(os.Args[2])
		if err != nil {
			exit(err.Error())
		}
		c := compare(v, w)
		fmt.Print(c)
	}
}

func exit(a ...interface{}) {
	fmt.Fprintln(os.Stderr, a...)
	os.Exit(1)
}
