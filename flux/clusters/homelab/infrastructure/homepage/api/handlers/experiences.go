package handlers

import (
	"net/http"
	"time"

	"github.com/brunovlucena/homelab/homepage-api/metrics"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// Experience represents work experience in the database
type Experience struct {
	ID           int            `json:"id" gorm:"primaryKey"`
	Title        string         `json:"title"`
	Company      string         `json:"company"`
	StartDate    string         `json:"start_date"`
	EndDate      *string        `json:"end_date"`
	Current      bool           `json:"current"`
	Description  string         `json:"description"`
	Technologies pq.StringArray `json:"technologies" gorm:"type:text[]"`
	Order        int            `json:"order"`
	Active       bool           `json:"active"`
}

// TableName overrides the table name
func (Experience) TableName() string {
	return "experience"
}

// GetExperiences returns all active experiences ordered by order
func GetExperiences(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		if db == nil {
			metrics.RecordExperienceLoadError("database_unavailable")
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "database not available"})
			return
		}

		var experiences []Experience
		if err := db.Where("active = ?", true).
			Order("\"order\" DESC, id DESC").
			Find(&experiences).Error; err != nil {
			// Record metrics for error
			metrics.RecordExperienceLoadError("database_query_error")
			metrics.RecordDatabaseError("select", "experience")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Record success metrics
		metrics.RecordExperienceLoadSuccess()
		metrics.ExperienceLoadDuration.Observe(time.Since(start).Seconds())

		c.JSON(http.StatusOK, experiences)
	}
}

// GetExperience returns a single experience by ID
func GetExperience(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		if db == nil {
			metrics.RecordExperienceLoadError("database_unavailable")
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "database not available"})
			return
		}

		id := c.Param("id")
		var experience Experience
		if err := db.Where("active = ?", true).First(&experience, id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				metrics.RecordExperienceLoadError("not_found")
			} else {
				metrics.RecordExperienceLoadError("database_query_error")
				metrics.RecordDatabaseError("select", "experience")
			}
			c.JSON(http.StatusNotFound, gin.H{"error": "experience not found"})
			return
		}

		// Record success metrics
		metrics.RecordExperienceLoadSuccess()
		metrics.ExperienceLoadDuration.Observe(time.Since(start).Seconds())

		c.JSON(http.StatusOK, experience)
	}
}

// CreateExperience creates a new experience
func CreateExperience(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if db == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "database not available"})
			return
		}

		var experience Experience
		if err := c.ShouldBindJSON(&experience); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Set default values
		experience.Active = true

		if err := db.Create(&experience).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, experience)
	}
}

// UpdateExperience updates an existing experience
func UpdateExperience(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if db == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "database not available"})
			return
		}

		id := c.Param("id")
		var experience Experience
		if err := db.First(&experience, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "experience not found"})
			return
		}

		if err := c.ShouldBindJSON(&experience); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := db.Save(&experience).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, experience)
	}
}

// DeleteExperience soft deletes an experience by setting active to false
func DeleteExperience(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if db == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "database not available"})
			return
		}

		id := c.Param("id")
		if err := db.Model(&Experience{}).Where("id = ?", id).Update("active", false).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "experience deleted"})
	}
}
