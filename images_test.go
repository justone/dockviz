package main

import (
	"regexp"
	"testing"
)

type DotTest struct {
	json    string
	regexps []string
}

type ShortTest struct {
	json    string
	regexps []string
}

type TreeTest struct {
	json       string
	startImage string
	noTrunc    bool
	incr       bool
	regexps    []string
}

func Test_BadJSON(t *testing.T) {
	_, err := parseImagesJSON([]byte(` "VirtualSize": 662553464, "Size": 662553464, "RepoTags": [ "<none>:<none>" ], "ParentId": "", "Id": "4c1208b690c68af3476b437e7bc2bcc460f062bda2094d2d8f21a7e70368d358", "Created": 1386114144 }]`))

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

	dotTests := []DotTest{
		DotTest{
			json: `[{ "VirtualSize": 662553464, "Size": 662553464, "RepoTags": [ "<none>:<none>" ], "ParentId": "", "Id": "4c1208b690c68af3476b437e7bc2bcc460f062bda2094d2d8f21a7e70368d358", "Created": 1386114144 }]`,
			regexps: []string{
				`base -> "4c1208b690c6"`,
			},
		},
		DotTest{
			json: `[{ "VirtualSize": 662553464, "Size": 0, "RepoTags": [ "foo:latest" ], "ParentId": "735f5db5626147582d2ae3f2c87be8e5e697c088574c5faaf8d4d1bccab99470", "Id": "c87be8e5e697c735f5db5626147582d2ae3f2088574c5faaf8d4d1bccab99470", "Created": 1386142123 },{ "VirtualSize": 662553464, "Size": 0, "RepoTags": [ "<none>:<none>" ], "ParentId": "4c1208b690c68af3476b437e7bc2bcc460f062bda2094d2d8f21a7e70368d358", "Id": "735f5db5626147582d2ae3f2c87be8e5e697c088574c5faaf8d4d1bccab99470", "Created": 1386142123 },{ "VirtualSize": 662553464, "Size": 662553464, "RepoTags": [ "<none>:<none>" ], "ParentId": "", "Id": "4c1208b690c68af3476b437e7bc2bcc460f062bda2094d2d8f21a7e70368d358", "Created": 1386114144 }]`,
			regexps: []string{
				`base -> "4c1208b690c6"`,
				`"4c1208b690c6" -> "735f5db56261"`,
				`"c87be8e5e697" \[label="c87be8e5e697\\nfoo:latest"`,
			},
		},
	}

	for _, dotTest := range dotTests {
		im, _ := parseImagesJSON([]byte(dotTest.json))
		byParent := collectChildren(im)
		roots := collectRoots(im)

		// TODO: test start image limiting

		result := jsonToDot(roots, byParent)

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

func Test_Tree(t *testing.T) {
	treeJSON := `[{"VirtualSize":674553464,"Size":2000000,"RepoTags":["foo:latest"],"ParentId":"735f5db5626147582d2ae3f2c87be8e5e697c088574c5faaf8d4d1bccab99470","Id":"c87be8e5e697c735f5db5626147582d2ae3f2088574c5faaf8d4d1bccab99470","Created":1386142123},{"VirtualSize":682553464,"Size":20000000,"RepoTags":["<none>:<none>"],"ParentId":"4c1208b690c68af3476b437e7bc2bcc460f062bda2094d2d8f21a7e70368d358","Id":"626147582d2ae3735f5db5f2c87be8e5e697c088574c5faaf8d4d1bccab99470","Created":1386142123},{"VirtualSize":712553464,"Size":30000000,"RepoTags":["base:latest"],"ParentId":"626147582d2ae3735f5db5f2c87be8e5e697c088574c5faaf8d4d1bccab99470","Id":"574c5faaf8d4d1bccab994626147582d2ae3735f5db5f2c87be8e5e697c08870","Created":1386142123},{"VirtualSize":752553464,"Size":40000000,"RepoTags":["<none>:<none>"],"ParentId":"574c5faaf8d4d1bccab994626147582d2ae3735f5db5f2c87be8e5e697c08870","Id":"aaf8d4d1bccab994574c5f626147582d2ae3735f5db5f2c87be8e5e697c08870","Created":1386142123},{"VirtualSize":672553464,"Size":10000000,"RepoTags":["<none>:<none>"],"ParentId":"4c1208b690c68af3476b437e7bc2bcc460f062bda2094d2d8f21a7e70368d358","Id":"735f5db5626147582d2ae3f2c87be8e5e697c088574c5faaf8d4d1bccab99470","Created":1386142123},{"VirtualSize":662553464,"Size":662553464,"RepoTags":["<none>:<none>"],"ParentId":"","Id":"4c1208b690c68af3476b437e7bc2bcc460f062bda2094d2d8f21a7e70368d358","Created":1386114144}]`

	treeTests := []TreeTest{
		TreeTest{
			json:       treeJSON,
			startImage: "",
			noTrunc:    false,
			incr:       false,
			regexps: []string{
				`(?m)└─4c1208b690c6 Virtual Size: 662.6 MB`,
				`(?m)  └─735f5db56261 Virtual Size: 672.6 MB`,
				`(?m)    └─c87be8e5e697 Virtual Size: 674.6 MB Tags: foo:latest`,
			},
		},
		TreeTest{
			json:       treeJSON,
			startImage: "",
			noTrunc:    false,
			incr:       true,
			regexps: []string{
				`(?m)└─4c1208b690c6 Virtual Size: 662.6 MB`,
				`(?m)  └─735f5db56261 Virtual Size: 10.0 MB`,
				`(?m)    └─c87be8e5e697 Virtual Size: 2.0 MB Tags: foo:latest`,
			},
		},
		TreeTest{
			json:       treeJSON,
			startImage: "626147582d2a",
			noTrunc:    false,
			incr:       false,
			regexps: []string{
				`(?m)└─626147582d2a`,
				`(?m)  └─574c5faaf8d4`,
				`(?m)    └─aaf8d4d1bcca`,
			},
		},
		TreeTest{
			json:       treeJSON,
			startImage: "base:latest",
			noTrunc:    true,
			incr:       false,
			regexps: []string{
				`(?m)└─574c5faaf8d4d1bccab994626147582d2ae3735f5db5f2c87be8e5e697c08870`,
				`(?m)  └─aaf8d4d1bccab994574c5f626147582d2ae3735f5db5f2c87be8e5e697c08870`,
			},
		},
	}

	for _, treeTest := range treeTests {
		im, _ := parseImagesJSON([]byte(treeTest.json))
		byParent := collectChildren(im)
		var roots []Image
		if len(treeTest.startImage) > 0 {
			startImage, _ := findStartImage(treeTest.startImage, im)
			startImage.ParentId = ""
			roots = []Image{*startImage}
		} else {
			roots = collectRoots(im)
		}
		result := jsonToTree(roots, byParent, treeTest.noTrunc, treeTest.incr)

		for _, regexp := range compileRegexps(t, treeTest.regexps) {
			if !regexp.MatchString(result) {
				t.Fatalf("images tree content '%s' did not match regexp '%s'", result, regexp)
			}
		}
	}
}

func Test_Short(t *testing.T) {
	shortJSON := `[ { "VirtualSize": 662553464, "Size": 0, "RepoTags": [ "foo:latest" ], "ParentId": "735f5db5626147582d2ae3f2c87be8e5e697c088574c5faaf8d4d1bccab99470", "Id": "c87be8e5e697c735f5db5626147582d2ae3f2088574c5faaf8d4d1bccab99470", "Created": 1386142123 }, { "VirtualSize": 682553464, "Size": 0, "RepoTags": [ "foo:1.0" ], "ParentId": "4c1208b690c68af3476b437e7bc2bcc460f062bda2094d2d8f21a7e70368d358", "Id": "626147582d2ae3735f5db5f2c87be8e5e697c088574c5faaf8d4d1bccab99470", "Created": 1386142123 }, { "VirtualSize": 712553464, "Size": 0, "RepoTags": [ "foo:2.0" ], "ParentId": "626147582d2ae3735f5db5f2c87be8e5e697c088574c5faaf8d4d1bccab99470", "Id": "574c5faaf8d4d1bccab994626147582d2ae3735f5db5f2c87be8e5e697c08870", "Created": 1386142123 }, { "VirtualSize": 752553464, "Size": 0, "RepoTags": [ "private.repo.com:5000:latest" ], "ParentId": "574c5faaf8d4d1bccab994626147582d2ae3735f5db5f2c87be8e5e697c08870", "Id": "aaf8d4d1bccab994574c5f626147582d2ae3735f5db5f2c87be8e5e697c08870", "Created": 1386142123 }, { "VirtualSize": 662553464, "Size": 0, "RepoTags": [ "<none>:<none>" ], "ParentId": "4c1208b690c68af3476b437e7bc2bcc460f062bda2094d2d8f21a7e70368d358", "Id": "735f5db5626147582d2ae3f2c87be8e5e697c088574c5faaf8d4d1bccab99470", "Created": 1386142123 }, { "VirtualSize": 662553464, "Size": 662553464, "RepoTags": [ "<none>:<none>" ], "ParentId": "", "Id": "4c1208b690c68af3476b437e7bc2bcc460f062bda2094d2d8f21a7e70368d358", "Created": 1386114144 } ]`

	shortTests := []ShortTest{
		ShortTest{
			json: shortJSON,
			regexps: []string{
				`(?m)foo: latest, 1.0, 2.0`,
				`(?m)private.repo.com:5000: latest`,
			},
		},
	}

	for _, shortTest := range shortTests {
		im, _ := parseImagesJSON([]byte(shortTest.json))
		result := jsonToShort(im)

		for _, regexp := range compileRegexps(t, shortTest.regexps) {
			if !regexp.MatchString(result) {
				t.Fatalf("images short content '%s' did not match regexp '%s'", result, regexp)
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
