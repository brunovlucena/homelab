package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
)

// ContextBuilder builds context from PostgreSQL data for LLM prompts
type ContextBuilder struct {
	db *sql.DB
}

// PersonalContext represents structured data about Bruno
type PersonalContext struct {
	About      AboutInfo     `json:"about"`
	Skills     []SkillInfo   `json:"skills"`
	Experience []ExpInfo     `json:"experience"`
	Projects   []ProjectInfo `json:"projects"`
	Contact    ContactInfo   `json:"contact"`
}

type AboutInfo struct {
	Description string `json:"description"`
}

type SkillInfo struct {
	Name        string `json:"name"`
	Category    string `json:"category"`
	Proficiency int    `json:"proficiency"`
}

type ExpInfo struct {
	Title        string   `json:"title"`
	Company      string   `json:"company"`
	Period       string   `json:"period"`
	Current      bool     `json:"current"`
	Description  string   `json:"description"`
	Technologies []string `json:"technologies"`
}

type ProjectInfo struct {
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	Type         string   `json:"type"`
	Technologies []string `json:"technologies"`
	GithubURL    string   `json:"github_url"`
	LiveURL      string   `json:"live_url"`
	Featured     bool     `json:"featured"`
}

type ContactInfo struct {
	Email        string `json:"email"`
	Location     string `json:"location"`
	LinkedIn     string `json:"linkedin"`
	GitHub       string `json:"github"`
	Availability string `json:"availability"`
}

// NewContextBuilder creates a new context builder
func NewContextBuilder(db *sql.DB) *ContextBuilder {
	return &ContextBuilder{db: db}
}

// BuildContext creates context based on user query
func (cb *ContextBuilder) BuildContext(query string) (string, error) {
	log.Printf("🔍 Building context for query: %s", query)

	// Analyze query to determine what data to include
	context := &PersonalContext{}

	// Always include basic about info
	about, err := cb.getAboutInfo()
	if err != nil {
		log.Printf("⚠️ Error getting about info: %v", err)
	} else {
		context.About = about
	}

	// Always include contact info for contact-related queries
	if cb.isContactQuery(query) {
		contact, err := cb.getContactInfo()
		if err != nil {
			log.Printf("⚠️ Error getting contact info: %v", err)
		} else {
			context.Contact = contact
		}
	}

	// Include skills if query mentions skills, technologies, or capabilities
	if cb.isSkillsQuery(query) {
		skills, err := cb.getRelevantSkills(query)
		if err != nil {
			log.Printf("⚠️ Error getting skills: %v", err)
		} else {
			context.Skills = skills
		}
	}

	// Include experience if query mentions work, experience, or companies
	if cb.isExperienceQuery(query) {
		experience, err := cb.getRelevantExperience(query)
		if err != nil {
			log.Printf("⚠️ Error getting experience: %v", err)
		} else {
			context.Experience = experience
		}
	}

	// Include projects if query mentions projects, work, or specific technologies
	if cb.isProjectsQuery(query) {
		projects, err := cb.getRelevantProjects(query)
		if err != nil {
			log.Printf("⚠️ Error getting projects: %v", err)
		} else {
			context.Projects = projects
		}
	}

	// Convert to formatted string for LLM
	return cb.formatContextForLLM(context, query), nil
}

// Query analysis methods
func (cb *ContextBuilder) isContactQuery(query string) bool {
	contactKeywords := []string{"contact", "email", "reach", "hire", "available", "linkedin", "github"}
	return cb.containsKeywords(query, contactKeywords)
}

func (cb *ContextBuilder) isSkillsQuery(query string) bool {
	skillKeywords := []string{"skill", "technology", "tech", "stack", "tools", "languages", "kubernetes", "aws", "go", "python", "devops", "sre"}
	return cb.containsKeywords(query, skillKeywords)
}

func (cb *ContextBuilder) isExperienceQuery(query string) bool {
	expKeywords := []string{"experience", "work", "job", "career", "company", "role", "position", "background", "mobimeo", "notifi", "crealytics"}
	return cb.containsKeywords(query, expKeywords)
}

func (cb *ContextBuilder) isProjectsQuery(query string) bool {
	projectKeywords := []string{"project", "site", "github", "build", "created", "developed", "bruno site", "knative"}
	return cb.containsKeywords(query, projectKeywords)
}

func (cb *ContextBuilder) containsKeywords(query string, keywords []string) bool {
	queryLower := strings.ToLower(query)
	for _, keyword := range keywords {
		if strings.Contains(queryLower, keyword) {
			return true
		}
	}
	return false
}

// Data retrieval methods
func (cb *ContextBuilder) getAboutInfo() (AboutInfo, error) {
	var about AboutInfo

	// Check if database connection is available
	if cb.db == nil {
		return about, fmt.Errorf("database connection not available")
	}

	var valueJSON string
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := cb.db.QueryRowContext(ctx, "SELECT value FROM content WHERE key = 'about'").Scan(&valueJSON)
	if err != nil {
		return about, err
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(valueJSON), &data); err != nil {
		return about, err
	}

	if desc, ok := data["description"].(string); ok {
		about.Description = desc
	}

	return about, nil
}

func (cb *ContextBuilder) getContactInfo() (ContactInfo, error) {
	var contact ContactInfo

	// Check if database connection is available
	if cb.db == nil {
		return contact, fmt.Errorf("database connection not available")
	}

	var valueJSON string
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := cb.db.QueryRowContext(ctx, "SELECT value FROM content WHERE key = 'contact'").Scan(&valueJSON)
	if err != nil {
		return contact, err
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(valueJSON), &data); err != nil {
		return contact, err
	}

	if email, ok := data["email"].(string); ok {
		contact.Email = email
	}
	if location, ok := data["location"].(string); ok {
		contact.Location = location
	}
	if linkedin, ok := data["linkedin"].(string); ok {
		contact.LinkedIn = linkedin
	}
	if github, ok := data["github"].(string); ok {
		contact.GitHub = github
	}
	if availability, ok := data["availability"].(string); ok {
		contact.Availability = availability
	}

	return contact, nil
}

func (cb *ContextBuilder) getRelevantSkills(query string) ([]SkillInfo, error) {
	var skills []SkillInfo

	// Check if database connection is available
	if cb.db == nil {
		return skills, fmt.Errorf("database connection not available")
	}

	// Get all active skills, ordered by proficiency and category
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := cb.db.QueryContext(ctx, `
		SELECT name, category, proficiency 
		FROM skills 
		WHERE active = true 
		ORDER BY proficiency DESC, category, name
		LIMIT 20
	`)
	if err != nil {
		return skills, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Error closing rows: %v", err)
		}
	}()

	for rows.Next() {
		var skill SkillInfo
		err := rows.Scan(&skill.Name, &skill.Category, &skill.Proficiency)
		if err != nil {
			continue
		}
		skills = append(skills, skill)
	}

	return skills, nil
}

func (cb *ContextBuilder) getRelevantExperience(query string) ([]ExpInfo, error) {
	var experiences []ExpInfo

	// Check if database connection is available
	if cb.db == nil {
		return experiences, fmt.Errorf("database connection not available")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := cb.db.QueryContext(ctx, `
		SELECT title, company, 
			CASE 
				WHEN current = true THEN start_date::text || ' - Present'
				ELSE start_date::text || ' - ' || end_date::text
			END as period,
			current, description, technologies
		FROM experience 
		WHERE active = true 
		ORDER BY "order" DESC
	`)
	if err != nil {
		return experiences, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Error closing rows: %v", err)
		}
	}()

	for rows.Next() {
		var exp ExpInfo
		var techArray sql.NullString

		err := rows.Scan(&exp.Title, &exp.Company, &exp.Period, &exp.Current, &exp.Description, &techArray)
		if err != nil {
			continue
		}

		// Parse PostgreSQL array
		if techArray.Valid {
			exp.Technologies = cb.parsePostgreSQLArray(techArray.String)
		}

		experiences = append(experiences, exp)
	}

	return experiences, nil
}

func (cb *ContextBuilder) getRelevantProjects(query string) ([]ProjectInfo, error) {
	var projects []ProjectInfo

	// Check if database connection is available
	if cb.db == nil {
		return projects, fmt.Errorf("database connection not available")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := cb.db.QueryContext(ctx, `
		SELECT title, description, type, github_url, live_url, technologies, featured
		FROM projects 
		WHERE active = true 
		ORDER BY featured DESC, "order"
	`)
	if err != nil {
		return projects, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Error closing rows: %v", err)
		}
	}()

	for rows.Next() {
		var project ProjectInfo
		var techArray sql.NullString
		var githubURL, liveURL sql.NullString

		err := rows.Scan(&project.Title, &project.Description, &project.Type,
			&githubURL, &liveURL, &techArray, &project.Featured)
		if err != nil {
			continue
		}

		if githubURL.Valid {
			project.GithubURL = githubURL.String
		}
		if liveURL.Valid {
			project.LiveURL = liveURL.String
		}

		// Parse PostgreSQL array
		if techArray.Valid {
			project.Technologies = cb.parsePostgreSQLArray(techArray.String)
		}

		projects = append(projects, project)
	}

	return projects, nil
}

// Helper method to parse PostgreSQL arrays
func (cb *ContextBuilder) parsePostgreSQLArray(arrayStr string) []string {
	// Remove curly braces and split by comma
	arrayStr = strings.Trim(arrayStr, "{}")
	if arrayStr == "" {
		return []string{}
	}

	parts := strings.Split(arrayStr, ",")
	var result []string
	for _, part := range parts {
		result = append(result, strings.Trim(part, " \""))
	}
	return result
}

// Format context for LLM prompt
func (cb *ContextBuilder) formatContextForLLM(context *PersonalContext, query string) string {
	var builder strings.Builder

	builder.WriteString("SYSTEM: Answer questions directly with facts. NO greetings, NO introductions. Maximum 2 sentences.\n\n")

	// About section
	if context.About.Description != "" {
		builder.WriteString(fmt.Sprintf("ABOUT BRUNO:\n%s\n\n", context.About.Description))
	}

	// Contact information
	if context.Contact.Email != "" {
		builder.WriteString("CONTACT INFORMATION:\n")
		builder.WriteString(fmt.Sprintf("- Email: %s\n", context.Contact.Email))
		if context.Contact.Location != "" {
			builder.WriteString(fmt.Sprintf("- Location: %s\n", context.Contact.Location))
		}
		if context.Contact.LinkedIn != "" {
			builder.WriteString(fmt.Sprintf("- LinkedIn: %s\n", context.Contact.LinkedIn))
		}
		if context.Contact.GitHub != "" {
			builder.WriteString(fmt.Sprintf("- GitHub: %s\n", context.Contact.GitHub))
		}
		if context.Contact.Availability != "" {
			builder.WriteString(fmt.Sprintf("- Availability: %s\n", context.Contact.Availability))
		}
		builder.WriteString("\n")
	}

	// Skills section
	if len(context.Skills) > 0 {
		builder.WriteString("SKILLS & TECHNOLOGIES:\n")
		skillsByCategory := make(map[string][]SkillInfo)
		for _, skill := range context.Skills {
			skillsByCategory[skill.Category] = append(skillsByCategory[skill.Category], skill)
		}

		for category, skills := range skillsByCategory {
			builder.WriteString(fmt.Sprintf("- %s: ", category))
			var skillNames []string
			for _, skill := range skills {
				skillNames = append(skillNames, fmt.Sprintf("%s (%d/5)", skill.Name, skill.Proficiency))
			}
			builder.WriteString(strings.Join(skillNames, ", "))
			builder.WriteString("\n")
		}
		builder.WriteString("\n")
	}

	// Experience section
	if len(context.Experience) > 0 {
		builder.WriteString("PROFESSIONAL EXPERIENCE:\n")
		for _, exp := range context.Experience {
			builder.WriteString(fmt.Sprintf("- %s at %s (%s)\n", exp.Title, exp.Company, exp.Period))
			if len(exp.Technologies) > 0 {
				builder.WriteString(fmt.Sprintf("  Tech: %s\n", strings.Join(exp.Technologies, ", ")))
			}
		}
		builder.WriteString("\n")
	}

	// Projects section
	if len(context.Projects) > 0 {
		builder.WriteString("KEY PROJECTS:\n")
		for _, project := range context.Projects {
			builder.WriteString(fmt.Sprintf("- %s (%s)\n", project.Title, project.Type))
			if len(project.Technologies) > 0 {
				builder.WriteString(fmt.Sprintf("  Tech: %s\n", strings.Join(project.Technologies, ", ")))
			}
		}
		builder.WriteString("\n")
	}

	builder.WriteString("CRITICAL: Keep responses SHORT and DIRECT. Maximum 2-3 sentences only.\n\n")

	builder.WriteString(fmt.Sprintf("USER QUESTION: %s\n", query))

	return builder.String()
}
