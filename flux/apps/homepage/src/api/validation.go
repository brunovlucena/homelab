package main

import (
	"html"
	"net/mail"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"
)

// =============================================================================
// 🔒 VALIDATION CONSTANTS
// =============================================================================

const (
	// Field length limits
	MaxTitleLength            = 255
	MaxDescriptionLength      = 5000
	MaxShortDescriptionLength = 500
	MaxTypeLength             = 100
	MaxNameLength             = 100
	MaxCategoryLength         = 100
	MaxIconLength             = 100
	MaxURLLength              = 2048
	MaxEmailLength            = 255
	MaxLocationLength         = 255
	MaxAvailabilityLength     = 500
	MaxContentValueLength     = 10000
	MaxCompanyLength          = 255
	MaxCompanyDescLength      = 2000
	MaxTechnologyLength       = 100
	MaxTechnologiesCount      = 50
	MaxHighlightsCount        = 20
	MaxHighlightTextLength    = 500
)

// =============================================================================
// 🔒 VALIDATION ERROR TYPES
// =============================================================================

// ValidationError represents a validation error with field and message
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationResult holds the result of validation
type ValidationResult struct {
	IsValid bool              `json:"is_valid"`
	Errors  []ValidationError `json:"errors,omitempty"`
}

// NewValidationResult creates a new empty validation result
func NewValidationResult() *ValidationResult {
	return &ValidationResult{
		IsValid: true,
		Errors:  []ValidationError{},
	}
}

// AddError adds an error to the validation result
func (v *ValidationResult) AddError(field, message string) {
	v.IsValid = false
	v.Errors = append(v.Errors, ValidationError{
		Field:   field,
		Message: message,
	})
}

// GetFirstError returns the first error message or empty string
func (v *ValidationResult) GetFirstError() string {
	if len(v.Errors) > 0 {
		return v.Errors[0].Message
	}
	return ""
}

// =============================================================================
// 🔒 STRING SANITIZATION FUNCTIONS
// =============================================================================

// SanitizeString removes potentially dangerous characters and HTML entities
// This provides basic XSS protection for text fields
func SanitizeString(s string) string {
	// HTML escape the string to prevent XSS
	s = html.EscapeString(s)
	// Remove null bytes
	s = strings.ReplaceAll(s, "\x00", "")
	// Trim leading/trailing whitespace
	s = strings.TrimSpace(s)
	return s
}

// SanitizeHTML provides basic HTML sanitization by escaping HTML entities
// For content that shouldn't contain any HTML
func SanitizeHTML(s string) string {
	return html.EscapeString(s)
}

// StripHTMLTags removes all HTML tags from a string
// Use this for fields that should be plain text only
func StripHTMLTags(s string) string {
	// Remove HTML tags
	re := regexp.MustCompile(`<[^>]*>`)
	s = re.ReplaceAllString(s, "")
	// Decode HTML entities back to readable text
	s = html.UnescapeString(s)
	// Then re-escape to prevent XSS
	s = html.EscapeString(s)
	return strings.TrimSpace(s)
}

// =============================================================================
// 🔒 URL VALIDATION FUNCTIONS
// =============================================================================

// ValidateURL checks if a URL is valid and uses http/https scheme
// Returns empty string if valid, error message otherwise
func ValidateURL(urlStr string, allowEmpty bool) string {
	if urlStr == "" {
		if allowEmpty {
			return ""
		}
		return "URL is required"
	}

	if len(urlStr) > MaxURLLength {
		return "URL exceeds maximum length"
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "Invalid URL format"
	}

	// Check scheme
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return "URL must use http or https scheme"
	}

	// Check host
	if parsedURL.Host == "" {
		return "URL must have a valid host"
	}

	return ""
}

// ValidateGitHubURL validates a GitHub URL
func ValidateGitHubURL(urlStr string, allowEmpty bool) string {
	if urlStr == "" && allowEmpty {
		return ""
	}

	baseErr := ValidateURL(urlStr, allowEmpty)
	if baseErr != "" {
		return baseErr
	}

	parsedURL, _ := url.Parse(urlStr)
	if !strings.Contains(parsedURL.Host, "github.com") &&
		!strings.Contains(parsedURL.Host, "github.io") {
		return "URL must be a GitHub URL"
	}

	return ""
}

// ValidateLinkedInURL validates a LinkedIn URL
func ValidateLinkedInURL(urlStr string, allowEmpty bool) string {
	if urlStr == "" && allowEmpty {
		return ""
	}

	baseErr := ValidateURL(urlStr, allowEmpty)
	if baseErr != "" {
		return baseErr
	}

	parsedURL, _ := url.Parse(urlStr)
	if !strings.Contains(parsedURL.Host, "linkedin.com") {
		return "URL must be a LinkedIn URL"
	}

	return ""
}

// =============================================================================
// 🔒 EMAIL VALIDATION
// =============================================================================

// ValidateEmail checks if an email address is valid
func ValidateEmail(email string, allowEmpty bool) string {
	if email == "" {
		if allowEmpty {
			return ""
		}
		return "Email is required"
	}

	if len(email) > MaxEmailLength {
		return "Email exceeds maximum length"
	}

	_, err := mail.ParseAddress(email)
	if err != nil {
		return "Invalid email format"
	}

	return ""
}

// =============================================================================
// 🔒 DATE VALIDATION
// =============================================================================

// DateFormatRegex matches YYYY-MM format
var DateFormatRegex = regexp.MustCompile(`^\d{4}-(0[1-9]|1[0-2])$`)

// ValidateDateFormat validates a date string in YYYY-MM format
func ValidateDateFormat(dateStr string, fieldName string, allowEmpty bool) string {
	if dateStr == "" {
		if allowEmpty {
			return ""
		}
		return fieldName + " is required"
	}

	if !DateFormatRegex.MatchString(dateStr) {
		return fieldName + " must be in YYYY-MM format"
	}

	return ""
}

// =============================================================================
// 🔒 STRING LENGTH VALIDATION
// =============================================================================

// ValidateStringLength validates string length
func ValidateStringLength(s string, fieldName string, minLen, maxLen int, required bool) string {
	length := utf8.RuneCountInString(s)

	if required && length == 0 {
		return fieldName + " is required"
	}

	if length == 0 && !required {
		return ""
	}

	if length < minLen {
		return fieldName + " is too short"
	}

	if length > maxLen {
		return fieldName + " exceeds maximum length"
	}

	return ""
}

// =============================================================================
// 🔒 CONTENT TYPE VALIDATION
// =============================================================================

// ValidContentTypes defines allowed content types
var ValidContentTypes = map[string]bool{
	"about":       true,
	"contact":     true,
	"hero":        true,
	"services":    true,
	"footer":      true,
	"navigation":  true,
	"social":      true,
	"skills":      true,
	"experience":  true,
	"projects":    true,
	"testimonial": true,
	"blog":        true,
	"misc":        true,
}

// ValidateContentType checks if a content type is valid
func ValidateContentType(contentType string) string {
	if contentType == "" {
		return "Content type is required"
	}

	if len(contentType) > MaxTypeLength {
		return "Content type exceeds maximum length"
	}

	// Allow any alphanumeric content type with underscores/hyphens
	validTypeRegex := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`)
	if !validTypeRegex.MatchString(contentType) {
		return "Content type contains invalid characters"
	}

	return ""
}

// =============================================================================
// 🔒 PROJECT VALIDATION
// =============================================================================

// ValidateProject validates a Project struct
func ValidateProject(p *Project) *ValidationResult {
	result := NewValidationResult()

	// Title validation
	if err := ValidateStringLength(p.Title, "Title", 1, MaxTitleLength, true); err != "" {
		result.AddError("title", err)
	}

	// Description validation
	if err := ValidateStringLength(p.Description, "Description", 1, MaxDescriptionLength, true); err != "" {
		result.AddError("description", err)
	}

	// Type validation
	if err := ValidateStringLength(p.Type, "Type", 1, MaxTypeLength, true); err != "" {
		result.AddError("type", err)
	}

	// GitHub URL validation (optional)
	if p.GithubURL != "" {
		if err := ValidateGitHubURL(p.GithubURL, true); err != "" {
			result.AddError("github_url", err)
		}
	}

	// Live URL validation (optional)
	if p.LiveURL != "" {
		if err := ValidateURL(p.LiveURL, true); err != "" {
			result.AddError("live_url", err)
		}
	}

	// Technologies validation
	if len(p.Technologies) > MaxTechnologiesCount {
		result.AddError("technologies", "Too many technologies listed")
	}
	for i, tech := range p.Technologies {
		if len(tech) > MaxTechnologyLength {
			result.AddError("technologies", "Technology name at index "+string(rune(i))+" exceeds maximum length")
			break
		}
	}

	// Sanitize fields
	p.Title = SanitizeString(p.Title)
	p.Description = SanitizeString(p.Description)
	p.Type = SanitizeString(p.Type)
	p.ShortDescription = SanitizeString(p.ShortDescription)
	for i := range p.Technologies {
		p.Technologies[i] = SanitizeString(p.Technologies[i])
	}

	return result
}

// =============================================================================
// 🔒 SKILL VALIDATION
// =============================================================================

// ValidateSkill validates a Skill struct
func ValidateSkill(s *Skill) *ValidationResult {
	result := NewValidationResult()

	// Name validation
	if err := ValidateStringLength(s.Name, "Name", 1, MaxNameLength, true); err != "" {
		result.AddError("name", err)
	}

	// Category validation
	if err := ValidateStringLength(s.Category, "Category", 1, MaxCategoryLength, true); err != "" {
		result.AddError("category", err)
	}

	// Proficiency validation (1-5)
	if s.Proficiency < 1 || s.Proficiency > 5 {
		result.AddError("proficiency", "Proficiency must be between 1 and 5")
	}

	// Icon validation (optional)
	if s.Icon != "" && len(s.Icon) > MaxIconLength {
		result.AddError("icon", "Icon exceeds maximum length")
	}

	// Sanitize fields
	s.Name = SanitizeString(s.Name)
	s.Category = SanitizeString(s.Category)
	s.Icon = SanitizeString(s.Icon)

	return result
}

// =============================================================================
// 🔒 EXPERIENCE VALIDATION
// =============================================================================

// ValidateExperience validates an Experience struct
func ValidateExperience(e *Experience) *ValidationResult {
	result := NewValidationResult()

	// Title validation
	if err := ValidateStringLength(e.Title, "Title", 1, MaxTitleLength, true); err != "" {
		result.AddError("title", err)
	}

	// Company validation
	if err := ValidateStringLength(e.Company, "Company", 1, MaxCompanyLength, true); err != "" {
		result.AddError("company", err)
	}

	// Start date validation
	if err := ValidateDateFormat(e.StartDate, "Start date", false); err != "" {
		result.AddError("start_date", err)
	}

	// End date validation (required if not current)
	if !e.Current {
		if e.EndDate == nil || *e.EndDate == "" {
			result.AddError("end_date", "End date is required when not current position")
		} else if err := ValidateDateFormat(*e.EndDate, "End date", false); err != "" {
			result.AddError("end_date", err)
		}
	}

	// Description validation (optional but has max length)
	if err := ValidateStringLength(e.Description, "Description", 0, MaxDescriptionLength, false); err != "" {
		result.AddError("description", err)
	}

	// Company description validation (optional)
	if e.CompanyDescription != nil {
		if err := ValidateStringLength(*e.CompanyDescription, "Company description", 0, MaxCompanyDescLength, false); err != "" {
			result.AddError("company_description", err)
		}
	}

	// Technologies validation
	if len(e.Technologies) > MaxTechnologiesCount {
		result.AddError("technologies", "Too many technologies listed")
	}
	for i, tech := range e.Technologies {
		if len(tech) > MaxTechnologyLength {
			result.AddError("technologies", "Technology name at index "+string(rune(i))+" exceeds maximum length")
			break
		}
	}

	// Sanitize fields
	e.Title = SanitizeString(e.Title)
	e.Company = SanitizeString(e.Company)
	e.Description = SanitizeString(e.Description)
	if e.CompanyDescription != nil {
		sanitized := SanitizeString(*e.CompanyDescription)
		e.CompanyDescription = &sanitized
	}
	for i := range e.Technologies {
		e.Technologies[i] = SanitizeString(e.Technologies[i])
	}

	return result
}

// =============================================================================
// 🔒 CONTENT VALIDATION
// =============================================================================

// ValidateContent validates a Content struct
func ValidateContent(c *Content) *ValidationResult {
	result := NewValidationResult()

	// Type validation
	if err := ValidateContentType(c.Type); err != "" {
		result.AddError("type", err)
	}

	// Value validation
	if err := ValidateStringLength(c.Value, "Value", 0, MaxContentValueLength, false); err != "" {
		result.AddError("value", err)
	}

	// Sanitize fields
	c.Type = SanitizeString(c.Type)
	c.Value = SanitizeString(c.Value)

	return result
}

// =============================================================================
// 🔒 SITE CONFIG VALIDATION
// =============================================================================

// ValidateSiteConfig validates a SiteConfig struct
func ValidateSiteConfig(sc *SiteConfig) *ValidationResult {
	result := NewValidationResult()

	// Hero title validation
	if err := ValidateStringLength(sc.HeroTitle, "Hero title", 0, MaxTitleLength, false); err != "" {
		result.AddError("hero_title", err)
	}

	// Hero subtitle validation
	if err := ValidateStringLength(sc.HeroSubtitle, "Hero subtitle", 0, MaxShortDescriptionLength, false); err != "" {
		result.AddError("hero_subtitle", err)
	}

	// Resume title validation
	if err := ValidateStringLength(sc.ResumeTitle, "Resume title", 0, MaxTitleLength, false); err != "" {
		result.AddError("resume_title", err)
	}

	// Resume subtitle validation
	if err := ValidateStringLength(sc.ResumeSubtitle, "Resume subtitle", 0, MaxShortDescriptionLength, false); err != "" {
		result.AddError("resume_subtitle", err)
	}

	// About title validation
	if err := ValidateStringLength(sc.AboutTitle, "About title", 0, MaxTitleLength, false); err != "" {
		result.AddError("about_title", err)
	}

	// About description validation
	if err := ValidateStringLength(sc.AboutDescription, "About description", 0, MaxDescriptionLength, false); err != "" {
		result.AddError("about_description", err)
	}

	// Sanitize fields
	sc.HeroTitle = SanitizeString(sc.HeroTitle)
	sc.HeroSubtitle = SanitizeString(sc.HeroSubtitle)
	sc.ResumeTitle = SanitizeString(sc.ResumeTitle)
	sc.ResumeSubtitle = SanitizeString(sc.ResumeSubtitle)
	sc.AboutTitle = SanitizeString(sc.AboutTitle)
	sc.AboutDescription = SanitizeString(sc.AboutDescription)

	return result
}

// =============================================================================
// 🔒 ABOUT DATA VALIDATION
// =============================================================================

// ValidateAboutData validates an AboutData struct
func ValidateAboutData(ad *AboutData) *ValidationResult {
	result := NewValidationResult()

	// Description validation
	if err := ValidateStringLength(ad.Description, "Description", 0, MaxDescriptionLength, false); err != "" {
		result.AddError("description", err)
	}

	// Highlights validation
	if len(ad.Highlights) > MaxHighlightsCount {
		result.AddError("highlights", "Too many highlights")
	}

	for i, highlight := range ad.Highlights {
		if len(highlight.Icon) > MaxIconLength {
			result.AddError("highlights", "Icon at index "+string(rune(i))+" exceeds maximum length")
		}
		if len(highlight.Text) > MaxHighlightTextLength {
			result.AddError("highlights", "Text at index "+string(rune(i))+" exceeds maximum length")
		}
	}

	// Sanitize fields
	ad.Description = SanitizeString(ad.Description)
	for i := range ad.Highlights {
		ad.Highlights[i].Icon = SanitizeString(ad.Highlights[i].Icon)
		ad.Highlights[i].Text = SanitizeString(ad.Highlights[i].Text)
	}

	return result
}

// =============================================================================
// 🔒 CONTACT DATA VALIDATION
// =============================================================================

// ValidateContactData validates a ContactData struct
func ValidateContactData(cd *ContactData) *ValidationResult {
	result := NewValidationResult()

	// Email validation
	if cd.Email != "" {
		if err := ValidateEmail(cd.Email, true); err != "" {
			result.AddError("email", err)
		}
	}

	// Location validation
	if err := ValidateStringLength(cd.Location, "Location", 0, MaxLocationLength, false); err != "" {
		result.AddError("location", err)
	}

	// LinkedIn URL validation
	if cd.LinkedIn != "" {
		if err := ValidateLinkedInURL(cd.LinkedIn, true); err != "" {
			result.AddError("linkedin", err)
		}
	}

	// GitHub URL validation
	if cd.GitHub != "" {
		if err := ValidateGitHubURL(cd.GitHub, true); err != "" {
			result.AddError("github", err)
		}
	}

	// Availability validation
	if err := ValidateStringLength(cd.Availability, "Availability", 0, MaxAvailabilityLength, false); err != "" {
		result.AddError("availability", err)
	}

	// Sanitize fields
	cd.Email = strings.TrimSpace(cd.Email)
	cd.Location = SanitizeString(cd.Location)
	cd.Availability = SanitizeString(cd.Availability)

	return result
}
