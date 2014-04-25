package main

import (
	"regexp"
	"testing"
)

type RunTest struct {
	json    string
	regexps []string
}

func Test_BadJSON(t *testing.T) {
	_, err := parseJSON([]byte(` "VirtualSize": 662553464, "Size": 662553464, "RepoTags": [ "<none>:<none>" ], "ParentId": "", "Id": "4c1208b690c68af3476b437e7bc2bcc460f062bda2094d2d8f21a7e70368d358", "Created": 1386114144 }]`))

	if err == nil {
		t.Error("invalid json did not cause an error")
	}
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
		RunTest{
			json: `[{ "VirtualSize": 662553464, "Size": 0, "RepoTags": [ "foo:latest" ], "ParentId": "735f5db5626147582d2ae3f2c87be8e5e697c088574c5faaf8d4d1bccab99470", "Id": "c87be8e5e697c735f5db5626147582d2ae3f2088574c5faaf8d4d1bccab99470", "Created": 1386142123 },{ "VirtualSize": 662553464, "Size": 0, "RepoTags": [ "<none>:<none>" ], "ParentId": "4c1208b690c68af3476b437e7bc2bcc460f062bda2094d2d8f21a7e70368d358", "Id": "735f5db5626147582d2ae3f2c87be8e5e697c088574c5faaf8d4d1bccab99470", "Created": 1386142123 },{ "VirtualSize": 662553464, "Size": 662553464, "RepoTags": [ "<none>:<none>" ], "ParentId": "", "Id": "4c1208b690c68af3476b437e7bc2bcc460f062bda2094d2d8f21a7e70368d358", "Created": 1386114144 }]`,
			regexps: []string{
				`base -> "4c1208b690c6"`,
				`"4c1208b690c6" -> "735f5db56261"`,
				`"c87be8e5e697" \[label="c87be8e5e697\\nfoo:latest"`,
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
