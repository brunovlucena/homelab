package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// 🔒 STRING SANITIZATION TESTS
// =============================================================================

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Normal string",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "HTML injection attempt",
			input:    "<script>alert('xss')</script>",
			expected: "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;",
		},
		{
			name:     "HTML tags in content",
			input:    "<b>Bold</b> and <i>italic</i>",
			expected: "&lt;b&gt;Bold&lt;/b&gt; and &lt;i&gt;italic&lt;/i&gt;",
		},
		{
			name:     "Null bytes",
			input:    "Hello\x00World",
			expected: "HelloWorld",
		},
		{
			name:     "Leading/trailing whitespace",
			input:    "  Hello World  ",
			expected: "Hello World",
		},
		{
			name:     "SQL injection attempt",
			input:    "'; DROP TABLE users;--",
			expected: "&#39;; DROP TABLE users;--",
		},
		{
			name:     "Ampersand escaping",
			input:    "A & B",
			expected: "A &amp; B",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStripHTMLTags(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Basic HTML tags",
			input:    "<p>Hello</p>",
			expected: "Hello",
		},
		{
			name:     "Multiple tags",
			input:    "<div><p>Hello</p><span>World</span></div>",
			expected: "HelloWorld",
		},
		{
			name:     "Script tags",
			input:    "<script>alert('xss')</script>Hello",
			expected: "alert(&#39;xss&#39;)Hello",
		},
		{
			name:     "Plain text",
			input:    "No HTML here",
			expected: "No HTML here",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StripHTMLTags(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// =============================================================================
// 🔒 URL VALIDATION TESTS
// =============================================================================

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		allowEmpty bool
		expectErr  bool
	}{
		{
			name:       "Valid HTTPS URL",
			url:        "https://example.com",
			allowEmpty: false,
			expectErr:  false,
		},
		{
			name:       "Valid HTTP URL",
			url:        "http://example.com",
			allowEmpty: false,
			expectErr:  false,
		},
		{
			name:       "Valid URL with path",
			url:        "https://example.com/path/to/resource",
			allowEmpty: false,
			expectErr:  false,
		},
		{
			name:       "Valid URL with query params",
			url:        "https://example.com?param=value&other=123",
			allowEmpty: false,
			expectErr:  false,
		},
		{
			name:       "Empty URL - not allowed",
			url:        "",
			allowEmpty: false,
			expectErr:  true,
		},
		{
			name:       "Empty URL - allowed",
			url:        "",
			allowEmpty: true,
			expectErr:  false,
		},
		{
			name:       "Invalid scheme - FTP",
			url:        "ftp://example.com",
			allowEmpty: false,
			expectErr:  true,
		},
		{
			name:       "Invalid scheme - javascript",
			url:        "javascript:alert('xss')",
			allowEmpty: false,
			expectErr:  true,
		},
		{
			name:       "No scheme",
			url:        "example.com",
			allowEmpty: false,
			expectErr:  true,
		},
		{
			name:       "No host",
			url:        "https://",
			allowEmpty: false,
			expectErr:  true,
		},
		{
			name:       "URL too long",
			url:        "https://example.com/" + strings.Repeat("a", 2100),
			allowEmpty: false,
			expectErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateURL(tt.url, tt.allowEmpty)
			if tt.expectErr {
				assert.NotEmpty(t, result, "Expected error but got none")
			} else {
				assert.Empty(t, result, "Expected no error but got: %s", result)
			}
		})
	}
}

func TestValidateGitHubURL(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		allowEmpty bool
		expectErr  bool
	}{
		{
			name:       "Valid GitHub URL",
			url:        "https://github.com/user/repo",
			allowEmpty: false,
			expectErr:  false,
		},
		{
			name:       "Valid GitHub.io URL",
			url:        "https://user.github.io/project",
			allowEmpty: false,
			expectErr:  false,
		},
		{
			name:       "Non-GitHub URL",
			url:        "https://gitlab.com/user/repo",
			allowEmpty: false,
			expectErr:  true,
		},
		{
			name:       "Empty URL - allowed",
			url:        "",
			allowEmpty: true,
			expectErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateGitHubURL(tt.url, tt.allowEmpty)
			if tt.expectErr {
				assert.NotEmpty(t, result, "Expected error but got none")
			} else {
				assert.Empty(t, result, "Expected no error but got: %s", result)
			}
		})
	}
}

func TestValidateLinkedInURL(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		allowEmpty bool
		expectErr  bool
	}{
		{
			name:       "Valid LinkedIn URL",
			url:        "https://www.linkedin.com/in/username",
			allowEmpty: false,
			expectErr:  false,
		},
		{
			name:       "LinkedIn company URL",
			url:        "https://linkedin.com/company/name",
			allowEmpty: false,
			expectErr:  false,
		},
		{
			name:       "Non-LinkedIn URL",
			url:        "https://twitter.com/username",
			allowEmpty: false,
			expectErr:  true,
		},
		{
			name:       "Empty URL - allowed",
			url:        "",
			allowEmpty: true,
			expectErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateLinkedInURL(tt.url, tt.allowEmpty)
			if tt.expectErr {
				assert.NotEmpty(t, result, "Expected error but got none")
			} else {
				assert.Empty(t, result, "Expected no error but got: %s", result)
			}
		})
	}
}

// =============================================================================
// 🔒 EMAIL VALIDATION TESTS
// =============================================================================

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name       string
		email      string
		allowEmpty bool
		expectErr  bool
	}{
		{
			name:       "Valid email",
			email:      "test@example.com",
			allowEmpty: false,
			expectErr:  false,
		},
		{
			name:       "Valid email with subdomain",
			email:      "test@mail.example.com",
			allowEmpty: false,
			expectErr:  false,
		},
		{
			name:       "Valid email with plus",
			email:      "test+tag@example.com",
			allowEmpty: false,
			expectErr:  false,
		},
		{
			name:       "Invalid email - no @",
			email:      "testexample.com",
			allowEmpty: false,
			expectErr:  true,
		},
		{
			name:       "Invalid email - no domain",
			email:      "test@",
			allowEmpty: false,
			expectErr:  true,
		},
		{
			name:       "Empty email - not allowed",
			email:      "",
			allowEmpty: false,
			expectErr:  true,
		},
		{
			name:       "Empty email - allowed",
			email:      "",
			allowEmpty: true,
			expectErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateEmail(tt.email, tt.allowEmpty)
			if tt.expectErr {
				assert.NotEmpty(t, result, "Expected error but got none")
			} else {
				assert.Empty(t, result, "Expected no error but got: %s", result)
			}
		})
	}
}

// =============================================================================
// 🔒 DATE VALIDATION TESTS
// =============================================================================

func TestValidateDateFormat(t *testing.T) {
	tests := []struct {
		name       string
		date       string
		allowEmpty bool
		expectErr  bool
	}{
		{
			name:       "Valid date - January",
			date:       "2023-01",
			allowEmpty: false,
			expectErr:  false,
		},
		{
			name:       "Valid date - December",
			date:       "2023-12",
			allowEmpty: false,
			expectErr:  false,
		},
		{
			name:       "Invalid month - 00",
			date:       "2023-00",
			allowEmpty: false,
			expectErr:  true,
		},
		{
			name:       "Invalid month - 13",
			date:       "2023-13",
			allowEmpty: false,
			expectErr:  true,
		},
		{
			name:       "Invalid format - YYYY/MM",
			date:       "2023/01",
			allowEmpty: false,
			expectErr:  true,
		},
		{
			name:       "Invalid format - full date",
			date:       "2023-01-15",
			allowEmpty: false,
			expectErr:  true,
		},
		{
			name:       "Invalid format - only year",
			date:       "2023",
			allowEmpty: false,
			expectErr:  true,
		},
		{
			name:       "Empty date - not allowed",
			date:       "",
			allowEmpty: false,
			expectErr:  true,
		},
		{
			name:       "Empty date - allowed",
			date:       "",
			allowEmpty: true,
			expectErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateDateFormat(tt.date, "Date", tt.allowEmpty)
			if tt.expectErr {
				assert.NotEmpty(t, result, "Expected error but got none")
			} else {
				assert.Empty(t, result, "Expected no error but got: %s", result)
			}
		})
	}
}

// =============================================================================
// 🔒 STRING LENGTH VALIDATION TESTS
// =============================================================================

func TestValidateStringLength(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		fieldName string
		minLen    int
		maxLen    int
		required  bool
		expectErr bool
	}{
		{
			name:      "Valid string",
			input:     "Hello",
			fieldName: "Test",
			minLen:    1,
			maxLen:    100,
			required:  true,
			expectErr: false,
		},
		{
			name:      "Empty string - required",
			input:     "",
			fieldName: "Test",
			minLen:    1,
			maxLen:    100,
			required:  true,
			expectErr: true,
		},
		{
			name:      "Empty string - optional",
			input:     "",
			fieldName: "Test",
			minLen:    1,
			maxLen:    100,
			required:  false,
			expectErr: false,
		},
		{
			name:      "String too short",
			input:     "Hi",
			fieldName: "Test",
			minLen:    5,
			maxLen:    100,
			required:  true,
			expectErr: true,
		},
		{
			name:      "String too long",
			input:     strings.Repeat("a", 101),
			fieldName: "Test",
			minLen:    1,
			maxLen:    100,
			required:  true,
			expectErr: true,
		},
		{
			name:      "Unicode string - within limit",
			input:     "こんにちは", // 5 characters
			fieldName: "Test",
			minLen:    1,
			maxLen:    10,
			required:  true,
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateStringLength(tt.input, tt.fieldName, tt.minLen, tt.maxLen, tt.required)
			if tt.expectErr {
				assert.NotEmpty(t, result, "Expected error but got none")
			} else {
				assert.Empty(t, result, "Expected no error but got: %s", result)
			}
		})
	}
}

// =============================================================================
// 🔒 CONTENT TYPE VALIDATION TESTS
// =============================================================================

func TestValidateContentType(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		expectErr   bool
	}{
		{
			name:        "Valid type - about",
			contentType: "about",
			expectErr:   false,
		},
		{
			name:        "Valid type - custom",
			contentType: "my_custom_type",
			expectErr:   false,
		},
		{
			name:        "Valid type with hyphen",
			contentType: "my-custom-type",
			expectErr:   false,
		},
		{
			name:        "Empty type",
			contentType: "",
			expectErr:   true,
		},
		{
			name:        "Invalid characters - spaces",
			contentType: "my type",
			expectErr:   true,
		},
		{
			name:        "Invalid characters - special",
			contentType: "my@type",
			expectErr:   true,
		},
		{
			name:        "Starts with number",
			contentType: "1type",
			expectErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateContentType(tt.contentType)
			if tt.expectErr {
				assert.NotEmpty(t, result, "Expected error but got none")
			} else {
				assert.Empty(t, result, "Expected no error but got: %s", result)
			}
		})
	}
}

// =============================================================================
// 🔒 PROJECT VALIDATION TESTS
// =============================================================================

func TestValidateProject(t *testing.T) {
	t.Run("Valid project", func(t *testing.T) {
		project := &Project{
			Title:        "Test Project",
			Description:  "A test project description",
			Type:         "web",
			Technologies: []string{"Go", "React"},
			Active:       true,
		}
		result := ValidateProject(project)
		assert.True(t, result.IsValid)
		assert.Empty(t, result.Errors)
	})

	t.Run("Valid project with URLs", func(t *testing.T) {
		project := &Project{
			Title:        "Test Project",
			Description:  "A test project description",
			Type:         "web",
			GithubURL:    "https://github.com/user/repo",
			LiveURL:      "https://example.com",
			Technologies: []string{"Go", "React"},
			Active:       true,
		}
		result := ValidateProject(project)
		assert.True(t, result.IsValid)
		assert.Empty(t, result.Errors)
	})

	t.Run("Empty title", func(t *testing.T) {
		project := &Project{
			Title:       "",
			Description: "A test project description",
			Type:        "web",
		}
		result := ValidateProject(project)
		assert.False(t, result.IsValid)
		assert.NotEmpty(t, result.Errors)
	})

	t.Run("Title too long", func(t *testing.T) {
		project := &Project{
			Title:       strings.Repeat("a", 300),
			Description: "A test project description",
			Type:        "web",
		}
		result := ValidateProject(project)
		assert.False(t, result.IsValid)
	})

	t.Run("Invalid GitHub URL", func(t *testing.T) {
		project := &Project{
			Title:       "Test Project",
			Description: "A test project description",
			Type:        "web",
			GithubURL:   "https://gitlab.com/user/repo",
		}
		result := ValidateProject(project)
		assert.False(t, result.IsValid)
	})

	t.Run("XSS in title is sanitized", func(t *testing.T) {
		project := &Project{
			Title:       "<script>alert('xss')</script>Test",
			Description: "A test project description",
			Type:        "web",
		}
		result := ValidateProject(project)
		assert.True(t, result.IsValid)
		assert.Contains(t, project.Title, "&lt;script&gt;")
	})
}

// =============================================================================
// 🔒 SKILL VALIDATION TESTS
// =============================================================================

func TestValidateSkill(t *testing.T) {
	t.Run("Valid skill", func(t *testing.T) {
		skill := &Skill{
			Name:        "Go",
			Category:    "Backend",
			Proficiency: 5,
			Icon:        "go-icon",
		}
		result := ValidateSkill(skill)
		assert.True(t, result.IsValid)
		assert.Empty(t, result.Errors)
	})

	t.Run("Empty name", func(t *testing.T) {
		skill := &Skill{
			Name:        "",
			Category:    "Backend",
			Proficiency: 5,
		}
		result := ValidateSkill(skill)
		assert.False(t, result.IsValid)
	})

	t.Run("Invalid proficiency - too low", func(t *testing.T) {
		skill := &Skill{
			Name:        "Go",
			Category:    "Backend",
			Proficiency: 0,
		}
		result := ValidateSkill(skill)
		assert.False(t, result.IsValid)
	})

	t.Run("Invalid proficiency - too high", func(t *testing.T) {
		skill := &Skill{
			Name:        "Go",
			Category:    "Backend",
			Proficiency: 6,
		}
		result := ValidateSkill(skill)
		assert.False(t, result.IsValid)
	})
}

// =============================================================================
// 🔒 EXPERIENCE VALIDATION TESTS
// =============================================================================

func TestValidateExperience(t *testing.T) {
	t.Run("Valid experience - current position", func(t *testing.T) {
		exp := &Experience{
			Title:     "Software Engineer",
			Company:   "Test Company",
			StartDate: "2020-01",
			Current:   true,
		}
		result := ValidateExperience(exp)
		assert.True(t, result.IsValid)
		assert.Empty(t, result.Errors)
	})

	t.Run("Valid experience - past position", func(t *testing.T) {
		endDate := "2023-12"
		exp := &Experience{
			Title:     "Software Engineer",
			Company:   "Test Company",
			StartDate: "2020-01",
			EndDate:   &endDate,
			Current:   false,
		}
		result := ValidateExperience(exp)
		assert.True(t, result.IsValid)
		assert.Empty(t, result.Errors)
	})

	t.Run("Invalid date format", func(t *testing.T) {
		exp := &Experience{
			Title:     "Software Engineer",
			Company:   "Test Company",
			StartDate: "2020/01",
			Current:   true,
		}
		result := ValidateExperience(exp)
		assert.False(t, result.IsValid)
	})

	t.Run("Missing end date for past position", func(t *testing.T) {
		exp := &Experience{
			Title:     "Software Engineer",
			Company:   "Test Company",
			StartDate: "2020-01",
			Current:   false,
		}
		result := ValidateExperience(exp)
		assert.False(t, result.IsValid)
	})

	t.Run("Empty title", func(t *testing.T) {
		exp := &Experience{
			Title:     "",
			Company:   "Test Company",
			StartDate: "2020-01",
			Current:   true,
		}
		result := ValidateExperience(exp)
		assert.False(t, result.IsValid)
	})
}

// =============================================================================
// 🔒 CONTENT VALIDATION TESTS
// =============================================================================

func TestValidateContent(t *testing.T) {
	t.Run("Valid content", func(t *testing.T) {
		content := &Content{
			Type:  "about",
			Value: "This is some content",
		}
		result := ValidateContent(content)
		assert.True(t, result.IsValid)
		assert.Empty(t, result.Errors)
	})

	t.Run("Invalid type - special characters", func(t *testing.T) {
		content := &Content{
			Type:  "about@test",
			Value: "This is some content",
		}
		result := ValidateContent(content)
		assert.False(t, result.IsValid)
	})

	t.Run("Value too long", func(t *testing.T) {
		content := &Content{
			Type:  "about",
			Value: strings.Repeat("a", 10001),
		}
		result := ValidateContent(content)
		assert.False(t, result.IsValid)
	})
}

// =============================================================================
// 🔒 SITE CONFIG VALIDATION TESTS
// =============================================================================

func TestValidateSiteConfig(t *testing.T) {
	t.Run("Valid site config", func(t *testing.T) {
		config := &SiteConfig{
			HeroTitle:        "Welcome",
			HeroSubtitle:     "A great site",
			ResumeTitle:      "Resume",
			ResumeSubtitle:   "My experience",
			AboutTitle:       "About",
			AboutDescription: "About me",
		}
		result := ValidateSiteConfig(config)
		assert.True(t, result.IsValid)
		assert.Empty(t, result.Errors)
	})

	t.Run("Title too long", func(t *testing.T) {
		config := &SiteConfig{
			HeroTitle: strings.Repeat("a", 300),
		}
		result := ValidateSiteConfig(config)
		assert.False(t, result.IsValid)
	})

	t.Run("XSS in content is sanitized", func(t *testing.T) {
		config := &SiteConfig{
			HeroTitle: "<script>alert('xss')</script>Welcome",
		}
		result := ValidateSiteConfig(config)
		assert.True(t, result.IsValid)
		assert.Contains(t, config.HeroTitle, "&lt;script&gt;")
	})
}

// =============================================================================
// 🔒 ABOUT DATA VALIDATION TESTS
// =============================================================================

func TestValidateAboutData(t *testing.T) {
	t.Run("Valid about data", func(t *testing.T) {
		aboutData := &AboutData{
			Description: "About me description",
		}
		result := ValidateAboutData(aboutData)
		assert.True(t, result.IsValid)
		assert.Empty(t, result.Errors)
	})

	t.Run("Description too long", func(t *testing.T) {
		aboutData := &AboutData{
			Description: strings.Repeat("a", 5001),
		}
		result := ValidateAboutData(aboutData)
		assert.False(t, result.IsValid)
	})
}

// =============================================================================
// 🔒 CONTACT DATA VALIDATION TESTS
// =============================================================================

func TestValidateContactData(t *testing.T) {
	t.Run("Valid contact data", func(t *testing.T) {
		contactData := &ContactData{
			Email:        "test@example.com",
			Location:     "Test Location",
			LinkedIn:     "https://www.linkedin.com/in/username",
			GitHub:       "https://github.com/username",
			Availability: "Available",
		}
		result := ValidateContactData(contactData)
		assert.True(t, result.IsValid)
		assert.Empty(t, result.Errors)
	})

	t.Run("Invalid email", func(t *testing.T) {
		contactData := &ContactData{
			Email: "invalid-email",
		}
		result := ValidateContactData(contactData)
		assert.False(t, result.IsValid)
	})

	t.Run("Invalid LinkedIn URL", func(t *testing.T) {
		contactData := &ContactData{
			LinkedIn: "https://twitter.com/username",
		}
		result := ValidateContactData(contactData)
		assert.False(t, result.IsValid)
	})

	t.Run("Invalid GitHub URL", func(t *testing.T) {
		contactData := &ContactData{
			GitHub: "https://gitlab.com/username",
		}
		result := ValidateContactData(contactData)
		assert.False(t, result.IsValid)
	})

	t.Run("Empty contact data - allowed", func(t *testing.T) {
		contactData := &ContactData{}
		result := ValidateContactData(contactData)
		assert.True(t, result.IsValid)
	})
}

// =============================================================================
// 🔒 VALIDATION RESULT TESTS
// =============================================================================

func TestValidationResult(t *testing.T) {
	t.Run("New result is valid", func(t *testing.T) {
		result := NewValidationResult()
		assert.True(t, result.IsValid)
		assert.Empty(t, result.Errors)
	})

	t.Run("Adding error makes result invalid", func(t *testing.T) {
		result := NewValidationResult()
		result.AddError("field", "error message")
		assert.False(t, result.IsValid)
		assert.Len(t, result.Errors, 1)
	})

	t.Run("GetFirstError returns first error", func(t *testing.T) {
		result := NewValidationResult()
		result.AddError("field1", "first error")
		result.AddError("field2", "second error")
		assert.Equal(t, "first error", result.GetFirstError())
	})

	t.Run("GetFirstError returns empty string when no errors", func(t *testing.T) {
		result := NewValidationResult()
		assert.Empty(t, result.GetFirstError())
	})
}

// =============================================================================
// 🏃 BENCHMARK TESTS
// =============================================================================

func BenchmarkSanitizeString(b *testing.B) {
	input := "<script>alert('xss')</script>Hello World"
	for i := 0; i < b.N; i++ {
		SanitizeString(input)
	}
}

func BenchmarkValidateURL(b *testing.B) {
	url := "https://example.com/path/to/resource?param=value"
	for i := 0; i < b.N; i++ {
		ValidateURL(url, false)
	}
}

func BenchmarkValidateProject(b *testing.B) {
	project := &Project{
		Title:        "Test Project",
		Description:  "A test project description that is a bit longer to simulate real data",
		Type:         "web",
		GithubURL:    "https://github.com/user/repo",
		LiveURL:      "https://example.com",
		Technologies: []string{"Go", "React", "PostgreSQL", "Docker"},
		Active:       true,
	}
	for i := 0; i < b.N; i++ {
		ValidateProject(project)
	}
}

func BenchmarkValidateEmail(b *testing.B) {
	email := "test@example.com"
	for i := 0; i < b.N; i++ {
		ValidateEmail(email, false)
	}
}
