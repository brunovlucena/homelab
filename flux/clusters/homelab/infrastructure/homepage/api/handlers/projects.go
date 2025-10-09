package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// Project represents a project in the database
type Project struct {
	ID           int            `json:"id" gorm:"primaryKey"`
	Title        string         `json:"title"`
	Description  string         `json:"description"`
	Technologies pq.StringArray `json:"technologies" gorm:"type:text[]"`
	GithubURL    string         `json:"github_url"`
	LiveURL      string         `json:"live_url"`
	Type         string         `json:"type"`
	GithubActive bool           `json:"github_active"`
}

// GetProjects returns all projects
func GetProjects(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if db == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "database not available"})
			return
		}

		var projects []Project
		if err := db.Find(&projects).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, projects)
	}
}

// GetProject returns a single project by ID
func GetProject(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if db == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "database not available"})
			return
		}

		id := c.Param("id")
		var project Project
		if err := db.First(&project, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
			return
		}

		c.JSON(http.StatusOK, project)
	}
}

// CreateProject creates a new project
func CreateProject(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if db == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "database not available"})
			return
		}

		var project Project
		if err := c.ShouldBindJSON(&project); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := db.Create(&project).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, project)
	}
}

// UpdateProject updates an existing project
func UpdateProject(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if db == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "database not available"})
			return
		}

		id := c.Param("id")
		var project Project
		if err := db.First(&project, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
			return
		}

		if err := c.ShouldBindJSON(&project); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := db.Save(&project).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, project)
	}
}

// DeleteProject deletes a project
func DeleteProject(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if db == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "database not available"})
			return
		}

		id := c.Param("id")
		if err := db.Delete(&Project{}, id).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "project deleted"})
	}
}
