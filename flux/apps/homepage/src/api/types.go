package main

// =============================================================================
// 📋 DATA STRUCTURES
// =============================================================================

// 🎯 Project represents a project
type Project struct {
	ID               int      `json:"id"`
	Title            string   `json:"title"`
	Description      string   `json:"description"`
	ShortDescription string   `json:"short_description"`
	Type             string   `json:"type"`
	Icon             string   `json:"icon"`
	GithubURL        string   `json:"github_url"`
	LiveURL          string   `json:"live_url"`
	Technologies     []string `json:"technologies"`
	Active           bool     `json:"active"`
	GithubActive     bool     `json:"github_active"`
}

// 📄 Content represents dynamic content from database
type Content struct {
	ID    int    `json:"id"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

// 👤 AboutData represents about page information
type AboutData struct {
	Description string `json:"description"`
	Highlights  []struct {
		Icon string `json:"icon"`
		Text string `json:"text"`
	} `json:"highlights"`
}

// 📞 ContactData represents contact information
type ContactData struct {
	Email        string `json:"email"`
	Location     string `json:"location"`
	LinkedIn     string `json:"linkedin"`
	GitHub       string `json:"github"`
	Availability string `json:"availability"`
}

// 🛠️ Skill represents a technical skill
type Skill struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	Proficiency int    `json:"proficiency"`
	Icon        string `json:"icon"`
	Order       int    `json:"order"`
	Active      bool   `json:"active"`
}

// 💼 Experience represents work experience
type Experience struct {
	ID                 int      `json:"id"`
	Title              string   `json:"title"`
	Company            string   `json:"company"`
	StartDate          string   `json:"start_date"`
	EndDate            *string  `json:"end_date"`
	Current            bool     `json:"current"`
	CompanyDescription *string  `json:"company_description"`
	Description        string   `json:"description"`
	Technologies       []string `json:"technologies"`
	Order              int      `json:"order"`
	Active             bool     `json:"active"`
}

// ⚙️ SiteConfig represents dynamic site configuration from database
type SiteConfig struct {
	HeroTitle        string `json:"hero_title"`
	HeroSubtitle     string `json:"hero_subtitle"`
	ResumeTitle      string `json:"resume_title"`
	ResumeSubtitle   string `json:"resume_subtitle"`
	AboutTitle       string `json:"about_title"`
	AboutDescription string `json:"about_description"`
}
