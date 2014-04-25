package main

import (
	"regexp"
	"testing"
)

type RunTest struct {
	json    string
	regexps []string
}

func Test_Dot(t *testing.T) {
	allMatch := []string{
		"(?s)digraph docker {.*}",
		`(?m) base \[style=invisible\]`,
	}
	allRegex := compileRegexps(t, allMatch)

	dotTests := []RunTest{
		RunTest{
			json: `[{ "VirtualSize": 662553464, "Size": 662553464, "RepoTags": [ "<none>:<none>" ], "ParentId": "", "Id": "4c1208b690c68af3476b437e7bc2bcc460f062bda2094d2d8f21a7e70368d358", "Created": 1386114144 }]`,
			regexps: []string{
				`base -> "4c1208b690c6"`,
			},
		},
	}

	for _, dotTest := range dotTests {
		im, _ := parseJSON([]byte(dotTest.json))
		result := jsonToDot(im)

		for _, regexp := range allRegex {
			if !regexp.MatchString(result) {
				t.Fatalf("images dot content '%s' did not match regexp '%s'", result, regexp)
			}
		}

		for _, regexp := range compileRegexps(t, dotTest.regexps) {
			if !regexp.MatchString(result) {
				t.Fatalf("images dot content '%s' did not match regexp '%s'", result, regexp)
			}
		}
	}
}

func compileRegexps(t *testing.T, regexpStrings []string) []*regexp.Regexp {

	compiledRegexps := []*regexp.Regexp{}
	for _, regexpString := range regexpStrings {
		regexp, err := regexp.Compile(regexpString)
		if err != nil {
			t.Errorf("Error in regex string '%s': %s", regexpString, err)
		}
		compiledRegexps = append(compiledRegexps, regexp)
	}

	return compiledRegexps
}
