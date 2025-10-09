package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// TestGetSkills tests the GetSkills handler
func TestGetSkills(t *testing.T) {
	tests := []struct {
		name               string
		dbSetup            func() (*gorm.DB, sqlmock.Sqlmock)
		expectedStatusCode int
		expectedError      bool
		validateResponse   func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "success - returns multiple skills",
			dbSetup: func() (*gorm.DB, sqlmock.Sqlmock) {
				db, mock, _ := sqlmock.New()
				gormDB, _ := gorm.Open(postgres.New(postgres.Config{
					Conn: db,
				}), &gorm.Config{})

				rows := sqlmock.NewRows([]string{
					"id", "name", "category", "proficiency", "icon", "order", "active",
				}).
					AddRow(1, "Kubernetes", "DevOps", 90, "kubernetes-icon", 1, true).
					AddRow(2, "Go", "Programming", 85, "golang-icon", 2, true).
					AddRow(3, "Docker", "DevOps", 95, "docker-icon", 1, true)

				mock.ExpectQuery(`SELECT \* FROM "skills" WHERE active = \$1 ORDER BY "order" ASC, category ASC, id ASC`).
					WithArgs(true).
					WillReturnRows(rows)

				return gormDB, mock
			},
			expectedStatusCode: http.StatusOK,
			expectedError:      false,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var skills []Skill
				err := json.Unmarshal(w.Body.Bytes(), &skills)
				assert.NoError(t, err)
				assert.Len(t, skills, 3)
				assert.Equal(t, "Kubernetes", skills[0].Name)
				assert.Equal(t, "DevOps", skills[0].Category)
				assert.Equal(t, 90, skills[0].Proficiency)
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
					"id", "name", "category", "proficiency", "icon", "order", "active",
				})

				mock.ExpectQuery(`SELECT \* FROM "skills"`).
					WillReturnRows(rows)

				return gormDB, mock
			},
			expectedStatusCode: http.StatusOK,
			expectedError:      false,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var skills []Skill
				err := json.Unmarshal(w.Body.Bytes(), &skills)
				assert.NoError(t, err)
				assert.Len(t, skills, 0)
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
			db, mock := tt.dbSetup()
			if mock != nil {
				defer func() {
					err := mock.ExpectationsWereMet()
					assert.NoError(t, err)
				}()
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/skills", nil)

			handler := GetSkills(db)
			handler(c)

			assert.Equal(t, tt.expectedStatusCode, w.Code)
			if tt.validateResponse != nil {
				tt.validateResponse(t, w)
			}
		})
	}
}

// TestGetSkill tests the GetSkill handler for a single skill
func TestGetSkill(t *testing.T) {
	tests := []struct {
		name               string
		skillID            string
		dbSetup            func() (*gorm.DB, sqlmock.Sqlmock)
		expectedStatusCode int
		expectedError      bool
		validateResponse   func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:    "success - returns single skill",
			skillID: "1",
			dbSetup: func() (*gorm.DB, sqlmock.Sqlmock) {
				db, mock, _ := sqlmock.New()
				gormDB, _ := gorm.Open(postgres.New(postgres.Config{
					Conn: db,
				}), &gorm.Config{})

				rows := sqlmock.NewRows([]string{
					"id", "name", "category", "proficiency", "icon", "order", "active",
				}).
					AddRow(1, "Kubernetes", "DevOps", 90, "kubernetes-icon", 1, true)

				mock.ExpectQuery(`SELECT \* FROM "skills" WHERE active = \$1 AND "skills"."id" = \$2 ORDER BY "skills"."id" LIMIT \$3`).
					WithArgs(true, "1", 1).
					WillReturnRows(rows)

				return gormDB, mock
			},
			expectedStatusCode: http.StatusOK,
			expectedError:      false,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var skill Skill
				err := json.Unmarshal(w.Body.Bytes(), &skill)
				assert.NoError(t, err)
				assert.Equal(t, 1, skill.ID)
				assert.Equal(t, "Kubernetes", skill.Name)
				assert.Equal(t, "DevOps", skill.Category)
			},
		},
		{
			name:    "error - skill not found",
			skillID: "999",
			dbSetup: func() (*gorm.DB, sqlmock.Sqlmock) {
				db, mock, _ := sqlmock.New()
				gormDB, _ := gorm.Open(postgres.New(postgres.Config{
					Conn: db,
				}), &gorm.Config{})

				mock.ExpectQuery(`SELECT \* FROM "skills"`).
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
				assert.Equal(t, "skill not found", response["error"])
			},
		},
		{
			name:    "error - database unavailable",
			skillID: "1",
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
			db, mock := tt.dbSetup()
			if mock != nil {
				defer func() {
					err := mock.ExpectationsWereMet()
					assert.NoError(t, err)
				}()
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/skills/"+tt.skillID, nil)
			c.Params = gin.Params{
				{Key: "id", Value: tt.skillID},
			}

			handler := GetSkill(db)
			handler(c)

			assert.Equal(t, tt.expectedStatusCode, w.Code)
			if tt.validateResponse != nil {
				tt.validateResponse(t, w)
			}
		})
	}
}

// TestCreateSkill tests the CreateSkill handler
func TestCreateSkill(t *testing.T) {
	tests := []struct {
		name               string
		requestBody        interface{}
		dbSetup            func() (*gorm.DB, sqlmock.Sqlmock)
		expectedStatusCode int
		validateResponse   func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "success - creates skill",
			requestBody: map[string]interface{}{
				"name":        "Python",
				"category":    "Programming",
				"proficiency": 80,
				"icon":        "python-icon",
				"order":       5,
			},
			dbSetup: func() (*gorm.DB, sqlmock.Sqlmock) {
				db, mock, _ := sqlmock.New()
				gormDB, _ := gorm.Open(postgres.New(postgres.Config{
					Conn: db,
				}), &gorm.Config{})

				mock.ExpectBegin()
				mock.ExpectQuery(`INSERT INTO "skills"`).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				mock.ExpectCommit()

				return gormDB, mock
			},
			expectedStatusCode: http.StatusCreated,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var skill Skill
				err := json.Unmarshal(w.Body.Bytes(), &skill)
				assert.NoError(t, err)
				assert.Equal(t, "Python", skill.Name)
				assert.True(t, skill.Active)
			},
		},
		{
			name:        "error - invalid JSON",
			requestBody: "invalid json",
			dbSetup: func() (*gorm.DB, sqlmock.Sqlmock) {
				db, mock, _ := sqlmock.New()
				gormDB, _ := gorm.Open(postgres.New(postgres.Config{
					Conn: db,
				}), &gorm.Config{})
				return gormDB, mock
			},
			expectedStatusCode: http.StatusBadRequest,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["error"], "json")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := tt.dbSetup()
			if mock != nil {
				defer func() {
					_ = mock.ExpectationsWereMet()
				}()
			}

			bodyBytes, _ := json.Marshal(tt.requestBody)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/skills", bytes.NewBuffer(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")

			handler := CreateSkill(db)
			handler(c)

			assert.Equal(t, tt.expectedStatusCode, w.Code)
			if tt.validateResponse != nil {
				tt.validateResponse(t, w)
			}
		})
	}
}

// TestUpdateSkill tests the UpdateSkill handler
func TestUpdateSkill(t *testing.T) {
	tests := []struct {
		name               string
		skillID            string
		requestBody        interface{}
		dbSetup            func() (*gorm.DB, sqlmock.Sqlmock)
		expectedStatusCode int
		validateResponse   func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:    "success - updates skill",
			skillID: "1",
			requestBody: map[string]interface{}{
				"name":        "Kubernetes Updated",
				"category":    "DevOps",
				"proficiency": 95,
				"icon":        "k8s-icon",
				"order":       1,
			},
			dbSetup: func() (*gorm.DB, sqlmock.Sqlmock) {
				db, mock, _ := sqlmock.New()
				gormDB, _ := gorm.Open(postgres.New(postgres.Config{
					Conn: db,
				}), &gorm.Config{})

				rows := sqlmock.NewRows([]string{
					"id", "name", "category", "proficiency", "icon", "order", "active",
				}).
					AddRow(1, "Kubernetes", "DevOps", 90, "kubernetes-icon", 1, true)

				mock.ExpectQuery(`SELECT \* FROM "skills" WHERE "skills"."id" = \$1 ORDER BY "skills"."id" LIMIT \$2`).
					WithArgs("1", 1).
					WillReturnRows(rows)

				mock.ExpectBegin()
				mock.ExpectExec(`UPDATE "skills"`).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()

				return gormDB, mock
			},
			expectedStatusCode: http.StatusOK,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var skill Skill
				err := json.Unmarshal(w.Body.Bytes(), &skill)
				assert.NoError(t, err)
				assert.Equal(t, "Kubernetes Updated", skill.Name)
			},
		},
		{
			name:        "error - skill not found",
			skillID:     "999",
			requestBody: map[string]interface{}{},
			dbSetup: func() (*gorm.DB, sqlmock.Sqlmock) {
				db, mock, _ := sqlmock.New()
				gormDB, _ := gorm.Open(postgres.New(postgres.Config{
					Conn: db,
				}), &gorm.Config{})

				mock.ExpectQuery(`SELECT \* FROM "skills" WHERE "skills"."id" = \$1 ORDER BY "skills"."id" LIMIT \$2`).
					WithArgs("999", 1).
					WillReturnRows(sqlmock.NewRows([]string{"id"}))

				return gormDB, mock
			},
			expectedStatusCode: http.StatusNotFound,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "skill not found", response["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock := tt.dbSetup()
			if mock != nil {
				defer func() {
					_ = mock.ExpectationsWereMet()
				}()
			}

			bodyBytes, _ := json.Marshal(tt.requestBody)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/skills/"+tt.skillID, bytes.NewBuffer(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")
			c.Params = gin.Params{{Key: "id", Value: tt.skillID}}

			handler := UpdateSkill(db)
			handler(c)

			assert.Equal(t, tt.expectedStatusCode, w.Code)
			if tt.validateResponse != nil {
				tt.validateResponse(t, w)
			}
		})
	}
}

// TestDeleteSkill tests the DeleteSkill handler
func TestDeleteSkill(t *testing.T) {
	tests := []struct {
		name               string
		skillID            string
		dbSetup            func() (*gorm.DB, sqlmock.Sqlmock)
		expectedStatusCode int
		validateResponse   func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:    "success - deletes skill",
			skillID: "1",
			dbSetup: func() (*gorm.DB, sqlmock.Sqlmock) {
				db, mock, _ := sqlmock.New()
				gormDB, _ := gorm.Open(postgres.New(postgres.Config{
					Conn: db,
				}), &gorm.Config{})

				mock.ExpectBegin()
				mock.ExpectExec(`UPDATE "skills" SET "active"=\$1 WHERE id = \$2`).
					WithArgs(false, "1").
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()

				return gormDB, mock
			},
			expectedStatusCode: http.StatusOK,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "skill deleted", response["message"])
			},
		},
		{
			name:    "error - database unavailable",
			skillID: "1",
			dbSetup: func() (*gorm.DB, sqlmock.Sqlmock) {
				return nil, nil
			},
			expectedStatusCode: http.StatusServiceUnavailable,
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
			db, mock := tt.dbSetup()
			if mock != nil {
				defer func() {
					err := mock.ExpectationsWereMet()
					assert.NoError(t, err)
				}()
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodDelete, "/api/v1/skills/"+tt.skillID, nil)
			c.Params = gin.Params{{Key: "id", Value: tt.skillID}}

			handler := DeleteSkill(db)
			handler(c)

			assert.Equal(t, tt.expectedStatusCode, w.Code)
			if tt.validateResponse != nil {
				tt.validateResponse(t, w)
			}
		})
	}
}
