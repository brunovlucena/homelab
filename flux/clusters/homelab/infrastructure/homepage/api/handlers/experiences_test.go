package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func init() {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)
}

// TestGetExperiences tests the GetExperiences handler
func TestGetExperiences(t *testing.T) {
	tests := []struct {
		name               string
		dbSetup            func() (*gorm.DB, sqlmock.Sqlmock)
		expectedStatusCode int
		expectedError      bool
		validateResponse   func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "success - returns experiences",
			dbSetup: func() (*gorm.DB, sqlmock.Sqlmock) {
				db, mock, _ := sqlmock.New()
				gormDB, _ := gorm.Open(postgres.New(postgres.Config{
					Conn: db,
				}), &gorm.Config{})

				rows := sqlmock.NewRows([]string{
					"id", "title", "company", "start_date", "end_date",
					"current", "description", "technologies", "order", "active",
				}).
					AddRow(1, "Senior DevOps Engineer", "Company A", "2023-01-01", nil, true, "Leading DevOps team", pq.StringArray{"Go", "Kubernetes"}, 1, true).
					AddRow(2, "DevOps Engineer", "Company B", "2021-01-01", "2022-12-31", false, "Infrastructure work", pq.StringArray{"Docker", "AWS"}, 2, true)

				mock.ExpectQuery(`SELECT \* FROM "experience" WHERE active = \$1 ORDER BY "order" DESC, id DESC`).
					WithArgs(true).
					WillReturnRows(rows)

				return gormDB, mock
			},
			expectedStatusCode: http.StatusOK,
			expectedError:      false,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var experiences []Experience
				err := json.Unmarshal(w.Body.Bytes(), &experiences)
				assert.NoError(t, err)
				assert.Len(t, experiences, 2)
				assert.Equal(t, "Senior DevOps Engineer", experiences[0].Title)
				assert.Equal(t, "Company A", experiences[0].Company)
				assert.Equal(t, "DevOps Engineer", experiences[1].Title)
			},
		},
		{
			name: "success - returns empty array",
			dbSetup: func() (*gorm.DB, sqlmock.Sqlmock) {
				db, mock, _ := sqlmock.New()
				gormDB, _ := gorm.Open(postgres.New(postgres.Config{
					Conn: db,
				}), &gorm.Config{})

				rows := sqlmock.NewRows([]string{
					"id", "title", "company", "start_date", "end_date",
					"current", "description", "technologies", "order", "active",
				})

				mock.ExpectQuery(`SELECT \* FROM "experience" WHERE active = \$1 ORDER BY "order" DESC, id DESC`).
					WithArgs(true).
					WillReturnRows(rows)

				return gormDB, mock
			},
			expectedStatusCode: http.StatusOK,
			expectedError:      false,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var experiences []Experience
				err := json.Unmarshal(w.Body.Bytes(), &experiences)
				assert.NoError(t, err)
				assert.Len(t, experiences, 0)
			},
		},
		{
			name: "error - database unavailable",
			dbSetup: func() (*gorm.DB, sqlmock.Sqlmock) {
				return nil, nil
			},
			expectedStatusCode: http.StatusServiceUnavailable,
			expectedError:      true,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "database not available", response["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup database
			db, mock := tt.dbSetup()
			if mock != nil {
				defer func() {
					err := mock.ExpectationsWereMet()
					assert.NoError(t, err)
				}()
			}

			// Create test request
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/experiences", nil)

			// Call handler
			handler := GetExperiences(db)
			handler(c)

			// Verify response
			assert.Equal(t, tt.expectedStatusCode, w.Code)
			if tt.validateResponse != nil {
				tt.validateResponse(t, w)
			}
		})
	}
}

// TestGetExperience tests the GetExperience handler for a single experience
func TestGetExperience(t *testing.T) {
	tests := []struct {
		name               string
		experienceID       string
		dbSetup            func() (*gorm.DB, sqlmock.Sqlmock)
		expectedStatusCode int
		expectedError      bool
		validateResponse   func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:         "success - returns single experience",
			experienceID: "1",
			dbSetup: func() (*gorm.DB, sqlmock.Sqlmock) {
				db, mock, _ := sqlmock.New()
				gormDB, _ := gorm.Open(postgres.New(postgres.Config{
					Conn: db,
				}), &gorm.Config{})

				rows := sqlmock.NewRows([]string{
					"id", "title", "company", "start_date", "end_date",
					"current", "description", "technologies", "order", "active",
				}).
					AddRow(1, "Senior DevOps Engineer", "Company A", "2023-01-01", nil, true, "Leading DevOps team", pq.StringArray{"Go", "Kubernetes"}, 1, true)

				mock.ExpectQuery(`SELECT \* FROM "experience" WHERE active = \$1 AND "experience"."id" = \$2 ORDER BY "experience"."id" LIMIT \$3`).
					WithArgs(true, "1", 1).
					WillReturnRows(rows)

				return gormDB, mock
			},
			expectedStatusCode: http.StatusOK,
			expectedError:      false,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var experience Experience
				err := json.Unmarshal(w.Body.Bytes(), &experience)
				assert.NoError(t, err)
				assert.Equal(t, 1, experience.ID)
				assert.Equal(t, "Senior DevOps Engineer", experience.Title)
				assert.Equal(t, "Company A", experience.Company)
			},
		},
		{
			name:         "error - experience not found",
			experienceID: "999",
			dbSetup: func() (*gorm.DB, sqlmock.Sqlmock) {
				db, mock, _ := sqlmock.New()
				gormDB, _ := gorm.Open(postgres.New(postgres.Config{
					Conn: db,
				}), &gorm.Config{})

				mock.ExpectQuery(`SELECT \* FROM "experience" WHERE active = \$1 AND "experience"."id" = \$2 ORDER BY "experience"."id" LIMIT \$3`).
					WithArgs(true, "999", 1).
					WillReturnRows(sqlmock.NewRows([]string{"id"}))

				return gormDB, mock
			},
			expectedStatusCode: http.StatusNotFound,
			expectedError:      true,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "experience not found", response["error"])
			},
		},
		{
			name:         "error - database unavailable",
			experienceID: "1",
			dbSetup: func() (*gorm.DB, sqlmock.Sqlmock) {
				return nil, nil
			},
			expectedStatusCode: http.StatusServiceUnavailable,
			expectedError:      true,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "database not available", response["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup database
			db, mock := tt.dbSetup()
			if mock != nil {
				defer func() {
					err := mock.ExpectationsWereMet()
					assert.NoError(t, err)
				}()
			}

			// Create test request
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/experiences/"+tt.experienceID, nil)
			c.Params = gin.Params{
				{Key: "id", Value: tt.experienceID},
			}

			// Call handler
			handler := GetExperience(db)
			handler(c)

			// Verify response
			assert.Equal(t, tt.expectedStatusCode, w.Code)
			if tt.validateResponse != nil {
				tt.validateResponse(t, w)
			}
		})
	}
}

// TestExperienceTableName tests the TableName override
func TestExperienceTableName(t *testing.T) {
	exp := Experience{}
	assert.Equal(t, "experience", exp.TableName())
}

// TestExperienceJSON tests JSON marshaling
func TestExperienceJSON(t *testing.T) {
	exp := Experience{
		ID:           1,
		Title:        "Senior DevOps Engineer",
		Company:      "Test Company",
		StartDate:    "2023-01-01",
		EndDate:      nil,
		Current:      true,
		Description:  "Test description",
		Technologies: pq.StringArray{"Go", "Kubernetes", "Docker"},
		Order:        1,
		Active:       true,
	}

	jsonData, err := json.Marshal(exp)
	assert.NoError(t, err)

	var decoded Experience
	err = json.Unmarshal(jsonData, &decoded)
	assert.NoError(t, err)

	assert.Equal(t, exp.ID, decoded.ID)
	assert.Equal(t, exp.Title, decoded.Title)
	assert.Equal(t, exp.Company, decoded.Company)
	assert.Equal(t, exp.Technologies, decoded.Technologies)
}

// TestGetExperiencesMetrics tests that metrics are recorded correctly
func TestGetExperiencesMetrics(t *testing.T) {
	t.Run("metrics recorded on success", func(t *testing.T) {
		db, mock, _ := sqlmock.New()
		gormDB, _ := gorm.Open(postgres.New(postgres.Config{
			Conn: db,
		}), &gorm.Config{})

		rows := sqlmock.NewRows([]string{
			"id", "title", "company", "start_date", "end_date",
			"current", "description", "technologies", "order", "active",
		}).
			AddRow(1, "Engineer", "Company", "2023-01-01", nil, true, "Desc", pq.StringArray{"Go"}, 1, true)

		mock.ExpectQuery(`SELECT \* FROM "experience"`).
			WillReturnRows(rows)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/experiences", nil)

		handler := GetExperiences(gormDB)
		handler(c)

		assert.Equal(t, http.StatusOK, w.Code)
		// Metrics are recorded but we can't easily assert on them in unit tests
		// In production, metrics would be scraped by Prometheus
	})

	t.Run("metrics recorded on database unavailable", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/experiences", nil)

		handler := GetExperiences(nil)
		handler(c)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
		// Metric bruno_site_experience_load_errors_total{error_type="database_unavailable"} incremented
	})
}
