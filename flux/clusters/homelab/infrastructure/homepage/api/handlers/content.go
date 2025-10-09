package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Content represents dynamic content in the database
type Content struct {
	ID    int             `json:"id" gorm:"primaryKey"`
	Key   string          `json:"key" gorm:"uniqueIndex"`
	Value json.RawMessage `json:"value" gorm:"type:jsonb"`
}

// TableName overrides the table name
func (Content) TableName() string {
	return "content"
}

// AboutData represents about page information
type AboutData struct {
	Description string `json:"description"`
}

// ContactData represents contact information
type ContactData struct {
	Email        string `json:"email"`
	Location     string `json:"location"`
	LinkedIn     string `json:"linkedin"`
	GitHub       string `json:"github"`
	Availability string `json:"availability"`
}

// GetContent returns all content
func GetContent(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if db == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "database not available"})
			return
		}

		var contents []Content
		if err := db.Find(&contents).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, contents)
	}
}

// GetContentByKey returns content by key (type)
func GetContentByKey(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if db == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "database not available"})
			return
		}

		key := c.Param("type")
		var content Content
		if err := db.Where("key = ?", key).First(&content).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "content not found"})
			return
		}

		c.JSON(http.StatusOK, content)
	}
}

// CreateContent creates new content
func CreateContent(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if db == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "database not available"})
			return
		}

		var content Content
		if err := c.ShouldBindJSON(&content); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := db.Create(&content).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, content)
	}
}

// UpdateContent updates existing content
func UpdateContent(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if db == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "database not available"})
			return
		}

		id := c.Param("id")
		var content Content
		if err := db.First(&content, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "content not found"})
			return
		}

		if err := c.ShouldBindJSON(&content); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := db.Save(&content).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, content)
	}
}

// DeleteContent deletes content
func DeleteContent(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if db == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "database not available"})
			return
		}

		id := c.Param("id")
		if err := db.Delete(&Content{}, id).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "content deleted"})
	}
}

// GetAbout returns the about content
func GetAbout(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if db == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "database not available"})
			return
		}

		var content Content
		if err := db.Where("key = ?", "about").First(&content).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "about content not found"})
			return
		}

		var aboutData AboutData
		if err := json.Unmarshal(content.Value, &aboutData); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse about data"})
			return
		}

		c.JSON(http.StatusOK, aboutData)
	}
}

// UpdateAbout updates the about content
func UpdateAbout(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if db == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "database not available"})
			return
		}

		var aboutData AboutData
		if err := c.ShouldBindJSON(&aboutData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		value, err := json.Marshal(aboutData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to encode about data"})
			return
		}

		var content Content
		result := db.Where("key = ?", "about").First(&content)
		if result.Error != nil {
			// Create new record
			content = Content{
				Key:   "about",
				Value: value,
			}
			if err := db.Create(&content).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		} else {
			// Update existing record
			content.Value = value
			if err := db.Save(&content).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}

		c.JSON(http.StatusOK, aboutData)
	}
}

// GetContact returns the contact content
func GetContact(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if db == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "database not available"})
			return
		}

		var content Content
		if err := db.Where("key = ?", "contact").First(&content).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "contact content not found"})
			return
		}

		var contactData ContactData
		if err := json.Unmarshal(content.Value, &contactData); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse contact data"})
			return
		}

		c.JSON(http.StatusOK, contactData)
	}
}

// UpdateContact updates the contact content
func UpdateContact(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if db == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "database not available"})
			return
		}

		var contactData ContactData
		if err := c.ShouldBindJSON(&contactData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		value, err := json.Marshal(contactData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to encode contact data"})
			return
		}

		var content Content
		result := db.Where("key = ?", "contact").First(&content)
		if result.Error != nil {
			// Create new record
			content = Content{
				Key:   "contact",
				Value: value,
			}
			if err := db.Create(&content).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		} else {
			// Update existing record
			content.Value = value
			if err := db.Save(&content).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}

		c.JSON(http.StatusOK, contactData)
	}
}
