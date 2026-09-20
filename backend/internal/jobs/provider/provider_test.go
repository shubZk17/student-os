package provider

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func serve(t *testing.T, body string) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv.URL + "/%s"
}

func TestFetchNormalizesEachSource(t *testing.T) {
	ctx, client := context.Background(), http.DefaultClient

	greenhouseURL = serve(t, `{"jobs":[{"id":42,"title":"Software Engineer Intern","absolute_url":"https://gh/42",
		"company_name":"Figma","content":"&lt;p&gt;Use Go &amp;amp; SQL&lt;/p&gt;&lt;p&gt;Remote OK&lt;/p&gt;",
		"first_published":"2026-01-28T18:57:29-05:00","application_deadline":"2026-10-01","location":{"name":"Remote - US"}}]}`)
	gh, err := Fetch(ctx, client, Board{Source: "greenhouse", Slug: "figma"})
	if err != nil {
		t.Fatal(err)
	}
	g := gh[0]
	if g.ExternalID != "figma:42" || g.CompanyName != "Figma" || g.Type != "INTERNSHIP" || !g.IsRemote ||
		g.Description != "Use Go & SQL\n\nRemote OK" || g.Deadline == nil || g.PostedAt.Year() != 2026 {
		t.Errorf("greenhouse: %+v", g)
	}

	leverURL = serve(t, `[{"id":"abc","text":"Backend Engineer","hostedUrl":"https://lv/abc","descriptionPlain":"Java",
		"createdAt":1782214185805,"workplaceType":"hybrid","categories":{"location":"London","commitment":"Intern"}}]`)
	lv, err := Fetch(ctx, client, Board{Source: "lever", Slug: "spotify", Company: "Spotify"})
	if err != nil {
		t.Fatal(err)
	}
	if l := lv[0]; l.ExternalID != "spotify:abc" || l.CompanyName != "Spotify" || l.Type != "INTERNSHIP" || l.IsRemote || l.Location != "London" {
		t.Errorf("lever: %+v", l)
	}

	ashbyURL = serve(t, `{"jobs":[
		{"id":"u1","title":"New Grad Engineer","jobUrl":"https://as/u1","descriptionPlain":"x","publishedAt":"2026-08-24T14:44:49.699+00:00","location":"","isRemote":true,"isListed":true,"employmentType":"FullTime"},
		{"id":"u2","title":"Hidden","isListed":false}]}`)
	as, err := Fetch(ctx, client, Board{Source: "ashby", Slug: "notion", Company: "Notion"})
	if err != nil {
		t.Fatal(err)
	}
	if len(as) != 1 || as[0].Type != "JOB" || as[0].Location != "Unspecified" || !as[0].IsRemote {
		t.Errorf("ashby: %+v", as)
	}

	greenhouseURL = serve(t, `not json`)
	if _, err := Fetch(ctx, client, Board{Source: "greenhouse", Slug: "x"}); err == nil {
		t.Error("expected error on bad payload")
	}
}

func TestIsStudentRole(t *testing.T) {
	for title, want := range map[string]bool{
		"Software Engineer Intern":          true,
		"New Grad Software Engineer, 2027":  true,
		"University Graduate - Backend":     true,
		"SDE-1":                             true,
		"Senior Software Engineer":          false,
		"University Recruiter":              false,
		"International Payments Engineer":   false,
		"Engineering Manager, Early Career": false,
	} {
		if got := IsStudentRole(title); got != want {
			t.Errorf("IsStudentRole(%q) = %v, want %v", title, got, want)
		}
	}
}

func TestSkillMatcher(t *testing.T) {
	m := NewSkillMatcher([]string{"Go", "C++", "Java", "Node.js", "SQL"})
	got := m.Match("Experience with C++, Node.js. and SQL; JavaScript a plus. Good to go!")
	// "Go" matches case-insensitively ("go!"); "Java" must not match inside "JavaScript".
	if want := []string{"Go", "C++", "Node.js", "SQL"}; !reflect.DeepEqual(got, want) {
		t.Errorf("Match = %v, want %v", got, want)
	}
}
