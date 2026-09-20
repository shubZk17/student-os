package authz

import "testing"

const (
	alice = "11111111-1111-1111-1111-111111111111"
	bob   = "22222222-2222-2222-2222-222222222222"
)

func newEngine(t *testing.T) *Engine {
	t.Helper()
	e, err := New()
	if err != nil {
		t.Fatalf("policy failed to parse: %v", err)
	}
	return e
}

// The point of the policy: a student reaches their own records and nobody else's.
// These are the IDOR cases — changing an ID in a URL must not grant access.
func TestOwnerAllowedStrangerDenied(t *testing.T) {
	e := newEngine(t)

	cases := []struct {
		action string
		typ    string
	}{
		{ViewApplication, TypeApplication},
		{UpdateApplication, TypeApplication},
		{DeleteApplication, TypeApplication},
		{ViewProject, TypeProject},
		{DeleteProject, TypeProject},
		{ViewNotification, TypeNotification},
		{MarkNotificationRead, TypeNotification},
		{ViewProfile, TypeProfile},
		{UpdateProfile, TypeProfile},
	}

	for _, tc := range cases {
		t.Run(tc.action, func(t *testing.T) {
			owned := Resource{Type: tc.typ, ID: "record-1", OwnerID: alice}

			if !e.Can(alice, tc.action, owned) {
				t.Errorf("owner was denied %s on their own %s", tc.action, tc.typ)
			}
			if e.Can(bob, tc.action, owned) {
				t.Errorf("SECURITY: %s allowed %s on another student's %s", bob, tc.action, tc.typ)
			}
		})
	}
}

// Opportunities are a public catalogue, so any authenticated student may read one
// even though it has no owner.
func TestOpportunityIsPublic(t *testing.T) {
	e := newEngine(t)
	opp := Resource{Type: TypeOpportunity, ID: "opp-1"}

	if !e.Can(alice, ViewOpportunity, opp) {
		t.Error("authenticated student denied a public opportunity")
	}
	if !e.Can(bob, ViewOpportunity, opp) {
		t.Error("authenticated student denied a public opportunity")
	}
}

// Cedar is deny-by-default. An action nobody wrote a policy for must not slip
// through just because the student owns the record.
func TestUnknownActionDenied(t *testing.T) {
	e := newEngine(t)
	owned := Resource{Type: TypeApplication, ID: "record-1", OwnerID: alice}

	if e.Can(alice, "deleteAllApplications", owned) {
		t.Error("SECURITY: an action with no policy was allowed")
	}
}

// A record with no owner attribute must not match an ownership policy. This guards
// the case where a handler forgets to populate OwnerID.
func TestMissingOwnerDenied(t *testing.T) {
	e := newEngine(t)
	ownerless := Resource{Type: TypeApplication, ID: "record-1"}

	if e.Can(alice, ViewApplication, ownerless) {
		t.Error("SECURITY: a record with no owner was treated as owned")
	}
}

// An empty student ID must never authorise anything, in case a handler is reached
// without an authenticated principal.
func TestEmptyPrincipalDenied(t *testing.T) {
	e := newEngine(t)

	if e.Can("", ViewApplication, Resource{Type: TypeApplication, ID: "r", OwnerID: alice}) {
		t.Error("SECURITY: empty principal was authorised")
	}
	if e.Can("", ViewApplication, Resource{Type: TypeApplication, ID: "r"}) {
		t.Error("SECURITY: empty principal matched an ownerless record")
	}
}
