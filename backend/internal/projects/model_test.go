package projects

import (
	"testing"

	"github.com/gin-gonic/gin/binding"
)

func TestProjectURLsMustBeHTTP(t *testing.T) {
	req := CreateProjectRequest{Title: "t", Description: "d", GithubURL: "javascript:alert(1)"}
	if binding.Validator.ValidateStruct(&req) == nil {
		t.Error("javascript: URL should be rejected")
	}
	req.GithubURL = "https://github.com/x/y"
	if err := binding.Validator.ValidateStruct(&req); err != nil {
		t.Errorf("https URL rejected: %v", err)
	}
}
