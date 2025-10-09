package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Skill represents a technical skill in the database
type Skill struct {
	ID          int    `json:"id" gorm:"primaryKey"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	Proficiency int    `json:"proficiency"`
	Icon        string `json:"icon"`
	Order       int    `json:"order"`
	Active      bool   `json:"active"`
}

// GetSkills returns all active skills ordered by order and category
func GetSkills(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if db == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "database not available"})
			return
		}

		var skills []Skill
		if err := db.Where("active = ?", true).
			Order("\"order\" ASC, category ASC, id ASC").
			Find(&skills).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, skills)
	}
}

// GetSkill returns a single skill by ID
func GetSkill(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if db == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "database not available"})
			return
		}

		id := c.Param("id")
		var skill Skill
		if err := db.Where("active = ?", true).First(&skill, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "skill not found"})
			return
		}

		c.JSON(http.StatusOK, skill)
	}
}

// CreateSkill creates a new skill
func CreateSkill(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if db == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "database not available"})
			return
		}

		var skill Skill
		if err := c.ShouldBindJSON(&skill); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Set default values
		skill.Active = true

		if err := db.Create(&skill).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, skill)
	}
}

// UpdateSkill updates an existing skill
func UpdateSkill(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if db == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "database not available"})
			return
		}

		id := c.Param("id")
		var skill Skill
		if err := db.First(&skill, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "skill not found"})
			return
		}

		if err := c.ShouldBindJSON(&skill); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := db.Save(&skill).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, skill)
	}
}

// DeleteSkill soft deletes a skill by setting active to false
func DeleteSkill(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if db == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "database not available"})
			return
		}

		id := c.Param("id")
		if err := db.Model(&Skill{}).Where("id = ?", id).Update("active", false).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "skill deleted"})
	}
}
