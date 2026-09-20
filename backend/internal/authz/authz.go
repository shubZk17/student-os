// Package authz answers "may this student do this to this record?" from a Cedar
// policy, so the rule lives in one readable, testable file instead of being implied
// by an `AND user_id = $n` repeated across every handler.
//
// This is a pre-check, not a replacement. Handlers keep their SQL ownership
// predicate: if a handler ever forgets to call Can, the database still refuses.
package authz

import (
	_ "embed"
	"fmt"

	cedar "github.com/cedar-policy/cedar-go"
	"github.com/cedar-policy/cedar-go/types"
)

//go:embed policies.cedar
var policyDocument []byte

// Action names, matching the Action:: identifiers in policies.cedar.
const (
	ViewApplication      = "viewApplication"
	UpdateApplication    = "updateApplication"
	DeleteApplication    = "deleteApplication"
	ViewProject          = "viewProject"
	CreateProject        = "createProject"
	DeleteProject        = "deleteProject"
	ViewNotification     = "viewNotification"
	MarkNotificationRead = "markNotificationRead"
	ViewProfile          = "viewProfile"
	UpdateProfile        = "updateProfile"
	ViewOpportunity      = "viewOpportunity"
)

// Resource types, matching the entity types used in requests.
const (
	TypeStudent      = "Student"
	TypeApplication  = "Application"
	TypeProject      = "Project"
	TypeNotification = "Notification"
	TypeProfile      = "Profile"
	TypeOpportunity  = "Opportunity"
)

// Engine evaluates the embedded policy. It is immutable and safe for concurrent use.
type Engine struct {
	policies *cedar.PolicySet
}

// New parses the embedded policy document. It fails loudly at startup rather than
// silently denying every request at runtime.
func New() (*Engine, error) {
	ps, err := cedar.NewPolicySetFromBytes("policies.cedar", policyDocument)
	if err != nil {
		return nil, fmt.Errorf("parse cedar policy: %w", err)
	}
	return &Engine{policies: ps}, nil
}

// Resource identifies a record and who owns it. ownerID is the student ID stored on
// the row; for public records such as opportunities it is empty.
type Resource struct {
	Type    string
	ID      string
	OwnerID string
}

// Can reports whether studentID may perform action on res.
//
// Cedar is deny-by-default: an unknown action or a resource whose owner does not
// match produces false.
func (e *Engine) Can(studentID, action string, res Resource) bool {
	principal := types.NewEntityUID(types.EntityType(TypeStudent), types.String(studentID))
	resourceUID := types.NewEntityUID(types.EntityType(res.Type), types.String(res.ID))

	attrs := types.RecordMap{}
	if res.OwnerID != "" {
		attrs["owner"] = types.NewEntityUID(types.EntityType(TypeStudent), types.String(res.OwnerID))
	}

	entities := types.EntityMap{
		resourceUID: types.Entity{
			UID:        resourceUID,
			Attributes: types.NewRecord(attrs),
		},
		principal: types.Entity{UID: principal},
	}

	decision, _ := cedar.Authorize(e.policies, entities, cedar.Request{
		Principal: principal,
		Action:    types.NewEntityUID("Action", types.String(action)),
		Resource:  resourceUID,
		Context:   types.NewRecord(types.RecordMap{}),
	})
	return decision == cedar.Allow
}
