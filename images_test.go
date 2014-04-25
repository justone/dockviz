package main

import (
	"testing"
)

func Test_Dot(t *testing.T) {
	json := `[{ "VirtualSize": 662553464, "Size": 662553464, "RepoTags": [ "<none>:<none>" ], "ParentId": "", "Id": "4c1208b690c68af3476b437e7bc2bcc460f062bda2094d2d8f21a7e70368d358", "Created": 1386114144 }]`

	expectedResult := `digraph docker {
 base -> "4c1208b690c6" [style=invis]
 base [style=invisible]
}
`

	im, _ := parseJSON([]byte(json))
	result := jsonToDot(im)

	if result == expectedResult {
		t.Log("Pass")
	} else {
		t.Errorf("|%s| and |%s| are different.", result, expectedResult)
	}
}
