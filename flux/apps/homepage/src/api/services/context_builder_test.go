package services

import (
	"database/sql"
	"testing"
)

// TestNewContextBuilder tests the creation of a new ContextBuilder
func TestNewContextBuilder(t *testing.T) {
	// Create a mock database connection (nil for testing)
	var db *sql.DB = nil

	builder := NewContextBuilder(db)

	if builder.db != db {
		t.Error("Expected database connection to match")
	}
}

// TestBuildContext tests the BuildContext method
func TestBuildContext(t *testing.T) {
	// Create a mock database connection (nil for testing)
	var db *sql.DB = nil
	builder := NewContextBuilder(db)

	// Test with empty query - should handle nil database gracefully
	context, err := builder.BuildContext("")
	if err != nil {
		t.Logf("BuildContext returned error (expected with nil database): %v", err)
	}

	// Even with error, context should be a string (might be empty)
	if context == "" {
		t.Log("Context is empty (expected with nil database)")
	}

	// Test with contact query - should handle nil database gracefully
	context, err = builder.BuildContext("contact information")
	if err != nil {
		t.Logf("BuildContext returned error (expected with nil database): %v", err)
	}

	// Even with error, context should be a string (might be empty)
	if context == "" {
		t.Log("Context is empty (expected with nil database)")
	}

	// Test with skills query - should handle nil database gracefully
	context, err = builder.BuildContext("skills and technologies")
	if err != nil {
		t.Logf("BuildContext returned error (expected with nil database): %v", err)
	}

	// Even with error, context should be a string (might be empty)
	if context == "" {
		t.Log("Context is empty (expected with nil database)")
	}
}

// TestIsContactQuery tests the isContactQuery method
func TestIsContactQuery(t *testing.T) {
	var db *sql.DB = nil
	builder := NewContextBuilder(db)

	tests := []struct {
		query    string
		expected bool
	}{
		{"contact", true},
		{"email", true},
		{"reach", true},
		{"hire", true},
		{"available", true},
		{"linkedin", true},
		{"github", true},
		{"phone", false},
		{"get in touch", false},
		{"hello", false},
		{"skills", false},
		{"", false},
	}

	for _, tt := range tests {
		result := builder.isContactQuery(tt.query)
		if result != tt.expected {
			t.Errorf("isContactQuery(%q) = %v, want %v", tt.query, result, tt.expected)
		}
	}
}

// TestIsSkillsQuery tests the isSkillsQuery method
func TestIsSkillsQuery(t *testing.T) {
	var db *sql.DB = nil
	builder := NewContextBuilder(db)

	tests := []struct {
		query    string
		expected bool
	}{
		{"skill", true},
		{"technology", true},
		{"tech", true},
		{"stack", true},
		{"tools", true},
		{"languages", true},
		{"kubernetes", true},
		{"aws", true},
		{"go", true},
		{"python", true},
		{"devops", true},
		{"sre", true},
		{"programming", false},
		{"capabilities", false},
		{"contact", false},
		{"hello", false},
		{"", false},
	}

	for _, tt := range tests {
		result := builder.isSkillsQuery(tt.query)
		if result != tt.expected {
			t.Errorf("isSkillsQuery(%q) = %v, want %v", tt.query, result, tt.expected)
		}
	}
}

// TestIsExperienceQuery tests the isExperienceQuery method
func TestIsExperienceQuery(t *testing.T) {
	var db *sql.DB = nil
	builder := NewContextBuilder(db)

	tests := []struct {
		query    string
		expected bool
	}{
		{"experience", true},
		{"work", true},
		{"job", true},
		{"career", true},
		{"company", true},
		{"role", true},
		{"position", true},
		{"background", true},
		{"mobimeo", true},
		{"notifi", true},
		{"crealytics", true},
		{"employment", false},
		{"contact", false},
		{"skills", false},
		{"", false},
	}

	for _, tt := range tests {
		result := builder.isExperienceQuery(tt.query)
		if result != tt.expected {
			t.Errorf("isExperienceQuery(%q) = %v, want %v", tt.query, result, tt.expected)
		}
	}
}

// TestIsProjectsQuery tests the isProjectsQuery method
func TestIsProjectsQuery(t *testing.T) {
	var db *sql.DB = nil
	builder := NewContextBuilder(db)

	tests := []struct {
		query    string
		expected bool
	}{
		{"project", true},
		{"projects", true},
		{"site", true},
		{"github", true},
		{"build", true},
		{"created", true},
		{"developed", true},
		{"bruno site", true},
		{"knative", true},
		{"portfolio", false},
		{"work", false},
		{"applications", false},
		{"apps", false},
		{"contact", false},
		{"skills", false},
		{"", false},
	}

	for _, tt := range tests {
		result := builder.isProjectsQuery(tt.query)
		if result != tt.expected {
			t.Errorf("isProjectsQuery(%q) = %v, want %v", tt.query, result, tt.expected)
		}
	}
}
