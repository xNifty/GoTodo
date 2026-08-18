package githubclient

import "testing"

func TestParseRepoFullName(t *testing.T) {
	cases := []struct {
		in          string
		owner, repo string
		ok          bool
	}{
		{"acme/widgets", "acme", "widgets", true},
		{" https://github.com/acme/widgets ", "acme", "widgets", true},
		{"https://github.com/acme/widgets.git", "acme", "widgets", true},
		{"github.com/acme/widgets", "acme", "widgets", true},
		{"", "", "", false},
		{"onlyowner", "", "", false},
	}
	for _, tc := range cases {
		o, r, err := ParseRepoFullName(tc.in)
		if tc.ok && err != nil {
			t.Fatalf("%q: unexpected err %v", tc.in, err)
		}
		if !tc.ok && err == nil {
			t.Fatalf("%q: expected error", tc.in)
		}
		if tc.ok && (o != tc.owner || r != tc.repo) {
			t.Fatalf("%q: got %s/%s want %s/%s", tc.in, o, r, tc.owner, tc.repo)
		}
	}
}

func TestParseIssueRef(t *testing.T) {
	o, r, n, err := ParseIssueRef("https://github.com/acme/widgets/issues/42")
	if err != nil || o != "acme" || r != "widgets" || n != 42 {
		t.Fatalf("url parse: %v %s/%s #%d", err, o, r, n)
	}
	o, r, n, err = ParseIssueRef("#7")
	if err != nil || o != "" || r != "" || n != 7 {
		t.Fatalf("number parse: %v %s/%s #%d", err, o, r, n)
	}
	if _, _, _, err := ParseIssueRef("nope"); err == nil {
		t.Fatal("expected error")
	}
}
