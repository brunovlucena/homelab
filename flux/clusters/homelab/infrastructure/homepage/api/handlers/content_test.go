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

// TestContentTableName tests the TableName override
func TestContentTableName(t *testing.T) {
	content := Content{}
	assert.Equal(t, "content", content.TableName())
}

// TestGetContent tests the GetContent handler
func TestGetContent(t *testing.T) {
	tests := []struct {
		name               string
		dbSetup            func() (*gorm.DB, sqlmock.Sqlmock)
		expectedStatusCode int
		validateResponse   func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "success - returns all content",
			dbSetup: func() (*gorm.DB, sqlmock.Sqlmock) {
				db, mock, _ := sqlmock.New()
				gormDB, _ := gorm.Open(postgres.New(postgres.Config{
					Conn: db,
				}), &gorm.Config{})

				aboutValue := json.RawMessage(`{"description":"Software Engineer"}`)
				contactValue := json.RawMessage(`{"email":"test@example.com"}`)

				rows := sqlmock.NewRows([]string{"id", "key", "value"}).
					AddRow(1, "about", aboutValue).
					AddRow(2, "contact", contactValue)

				mock.ExpectQuery(`SELECT \* FROM "content"`).
					WillReturnRows(rows)

				return gormDB, mock
			},
			expectedStatusCode: http.StatusOK,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var contents []Content
				err := json.Unmarshal(w.Body.Bytes(), &contents)
				assert.NoError(t, err)
				assert.Len(t, contents, 2)
				assert.Equal(t, "about", contents[0].Key)
			},
		},
		{
			name: "error - database unavailable",
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
			c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/content", nil)

			handler := GetContent(db)
			handler(c)

			assert.Equal(t, tt.expectedStatusCode, w.Code)
			if tt.validateResponse != nil {
				tt.validateResponse(t, w)
			}
		})
	}
}

// TestGetContentByKey tests the GetContentByKey handler
func TestGetContentByKey(t *testing.T) {
	tests := []struct {
		name               string
		contentKey         string
		dbSetup            func() (*gorm.DB, sqlmock.Sqlmock)
		expectedStatusCode int
		validateResponse   func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:       "success - returns content by key",
			contentKey: "about",
			dbSetup: func() (*gorm.DB, sqlmock.Sqlmock) {
				db, mock, _ := sqlmock.New()
				gormDB, _ := gorm.Open(postgres.New(postgres.Config{
					Conn: db,
				}), &gorm.Config{})

				aboutValue := json.RawMessage(`{"description":"Software Engineer"}`)

				rows := sqlmock.NewRows([]string{"id", "key", "value"}).
					AddRow(1, "about", aboutValue)

				mock.ExpectQuery(`SELECT \* FROM "content" WHERE key = \$1 ORDER BY "content"."id" LIMIT \$2`).
					WithArgs("about", 1).
					WillReturnRows(rows)

				return gormDB, mock
			},
			expectedStatusCode: http.StatusOK,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var content Content
				err := json.Unmarshal(w.Body.Bytes(), &content)
				assert.NoError(t, err)
				assert.Equal(t, "about", content.Key)
			},
		},
		{
			name:       "error - content not found",
			contentKey: "nonexistent",
			dbSetup: func() (*gorm.DB, sqlmock.Sqlmock) {
				db, mock, _ := sqlmock.New()
				gormDB, _ := gorm.Open(postgres.New(postgres.Config{
					Conn: db,
				}), &gorm.Config{})

				mock.ExpectQuery(`SELECT \* FROM "content"`).
					WithArgs("nonexistent", 1).
					WillReturnRows(sqlmock.NewRows([]string{"id"}))

				return gormDB, mock
			},
			expectedStatusCode: http.StatusNotFound,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "content not found", response["error"])
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
			c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/content/"+tt.contentKey, nil)
			c.Params = gin.Params{{Key: "type", Value: tt.contentKey}}

			handler := GetContentByKey(db)
			handler(c)

			assert.Equal(t, tt.expectedStatusCode, w.Code)
			if tt.validateResponse != nil {
				tt.validateResponse(t, w)
			}
		})
	}
}

// TestGetAbout tests the GetAbout handler
func TestGetAbout(t *testing.T) {
	tests := []struct {
		name               string
		dbSetup            func() (*gorm.DB, sqlmock.Sqlmock)
		expectedStatusCode int
		validateResponse   func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "success - returns about data",
			dbSetup: func() (*gorm.DB, sqlmock.Sqlmock) {
				db, mock, _ := sqlmock.New()
				gormDB, _ := gorm.Open(postgres.New(postgres.Config{
					Conn: db,
				}), &gorm.Config{})

				aboutValue := json.RawMessage(`{"description":"Experienced Software Engineer"}`)

				rows := sqlmock.NewRows([]string{"id", "key", "value"}).
					AddRow(1, "about", aboutValue)

				mock.ExpectQuery(`SELECT \* FROM "content" WHERE key = \$1 ORDER BY "content"."id" LIMIT \$2`).
					WithArgs("about", 1).
					WillReturnRows(rows)

				return gormDB, mock
			},
			expectedStatusCode: http.StatusOK,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var aboutData AboutData
				err := json.Unmarshal(w.Body.Bytes(), &aboutData)
				assert.NoError(t, err)
				assert.Equal(t, "Experienced Software Engineer", aboutData.Description)
			},
		},
		{
			name: "error - about content not found",
			dbSetup: func() (*gorm.DB, sqlmock.Sqlmock) {
				db, mock, _ := sqlmock.New()
				gormDB, _ := gorm.Open(postgres.New(postgres.Config{
					Conn: db,
				}), &gorm.Config{})

				mock.ExpectQuery(`SELECT \* FROM "content"`).
					WithArgs("about", 1).
					WillReturnRows(sqlmock.NewRows([]string{"id"}))

				return gormDB, mock
			},
			expectedStatusCode: http.StatusNotFound,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "about content not found", response["error"])
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
			c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/about", nil)

			handler := GetAbout(db)
			handler(c)

			assert.Equal(t, tt.expectedStatusCode, w.Code)
			if tt.validateResponse != nil {
				tt.validateResponse(t, w)
			}
		})
	}
}

// TestUpdateAbout tests the UpdateAbout handler
func TestUpdateAbout(t *testing.T) {
	tests := []struct {
		name               string
		requestBody        interface{}
		dbSetup            func() (*gorm.DB, sqlmock.Sqlmock)
		expectedStatusCode int
		validateResponse   func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "success - creates new about content",
			requestBody: AboutData{
				Description: "New description",
			},
			dbSetup: func() (*gorm.DB, sqlmock.Sqlmock) {
				db, mock, _ := sqlmock.New()
				gormDB, _ := gorm.Open(postgres.New(postgres.Config{
					Conn: db,
				}), &gorm.Config{})

				// First query returns not found
				mock.ExpectQuery(`SELECT \* FROM "content" WHERE key = \$1 ORDER BY "content"."id" LIMIT \$2`).
					WithArgs("about", 1).
					WillReturnRows(sqlmock.NewRows([]string{"id"}))

				// Then create new record
				mock.ExpectBegin()
				mock.ExpectQuery(`INSERT INTO "content"`).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				mock.ExpectCommit()

				return gormDB, mock
			},
			expectedStatusCode: http.StatusOK,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var aboutData AboutData
				err := json.Unmarshal(w.Body.Bytes(), &aboutData)
				assert.NoError(t, err)
				assert.Equal(t, "New description", aboutData.Description)
			},
		},
		{
			name: "success - updates existing about content",
			requestBody: AboutData{
				Description: "Updated description",
			},
			dbSetup: func() (*gorm.DB, sqlmock.Sqlmock) {
				db, mock, _ := sqlmock.New()
				gormDB, _ := gorm.Open(postgres.New(postgres.Config{
					Conn: db,
				}), &gorm.Config{})

				aboutValue := json.RawMessage(`{"description":"Old description"}`)

				// First query returns existing record
				rows := sqlmock.NewRows([]string{"id", "key", "value"}).
					AddRow(1, "about", aboutValue)

				mock.ExpectQuery(`SELECT \* FROM "content" WHERE key = \$1 ORDER BY "content"."id" LIMIT \$2`).
					WithArgs("about", 1).
					WillReturnRows(rows)

				// Then update record
				mock.ExpectBegin()
				mock.ExpectExec(`UPDATE "content"`).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()

				return gormDB, mock
			},
			expectedStatusCode: http.StatusOK,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var aboutData AboutData
				err := json.Unmarshal(w.Body.Bytes(), &aboutData)
				assert.NoError(t, err)
				assert.Equal(t, "Updated description", aboutData.Description)
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
			c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/about", bytes.NewBuffer(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")

			handler := UpdateAbout(db)
			handler(c)

			assert.Equal(t, tt.expectedStatusCode, w.Code)
			if tt.validateResponse != nil {
				tt.validateResponse(t, w)
			}
		})
	}
}

// TestGetContact tests the GetContact handler
func TestGetContact(t *testing.T) {
	tests := []struct {
		name               string
		dbSetup            func() (*gorm.DB, sqlmock.Sqlmock)
		expectedStatusCode int
		validateResponse   func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "success - returns contact data",
			dbSetup: func() (*gorm.DB, sqlmock.Sqlmock) {
				db, mock, _ := sqlmock.New()
				gormDB, _ := gorm.Open(postgres.New(postgres.Config{
					Conn: db,
				}), &gorm.Config{})

				contactValue := json.RawMessage(`{"email":"test@example.com","location":"San Francisco"}`)

				rows := sqlmock.NewRows([]string{"id", "key", "value"}).
					AddRow(1, "contact", contactValue)

				mock.ExpectQuery(`SELECT \* FROM "content" WHERE key = \$1 ORDER BY "content"."id" LIMIT \$2`).
					WithArgs("contact", 1).
					WillReturnRows(rows)

				return gormDB, mock
			},
			expectedStatusCode: http.StatusOK,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var contactData ContactData
				err := json.Unmarshal(w.Body.Bytes(), &contactData)
				assert.NoError(t, err)
				assert.Equal(t, "test@example.com", contactData.Email)
				assert.Equal(t, "San Francisco", contactData.Location)
			},
		},
		{
			name: "error - contact content not found",
			dbSetup: func() (*gorm.DB, sqlmock.Sqlmock) {
				db, mock, _ := sqlmock.New()
				gormDB, _ := gorm.Open(postgres.New(postgres.Config{
					Conn: db,
				}), &gorm.Config{})

				mock.ExpectQuery(`SELECT \* FROM "content"`).
					WithArgs("contact", 1).
					WillReturnRows(sqlmock.NewRows([]string{"id"}))

				return gormDB, mock
			},
			expectedStatusCode: http.StatusNotFound,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, "contact content not found", response["error"])
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
			c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/contact", nil)

			handler := GetContact(db)
			handler(c)

			assert.Equal(t, tt.expectedStatusCode, w.Code)
			if tt.validateResponse != nil {
				tt.validateResponse(t, w)
			}
		})
	}
}

// TestUpdateContact tests the UpdateContact handler
func TestUpdateContact(t *testing.T) {
	tests := []struct {
		name               string
		requestBody        interface{}
		dbSetup            func() (*gorm.DB, sqlmock.Sqlmock)
		expectedStatusCode int
		validateResponse   func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "success - creates new contact content",
			requestBody: ContactData{
				Email:        "new@example.com",
				Location:     "New York",
				LinkedIn:     "linkedin.com/in/test",
				GitHub:       "github.com/test",
				Availability: "Available",
			},
			dbSetup: func() (*gorm.DB, sqlmock.Sqlmock) {
				db, mock, _ := sqlmock.New()
				gormDB, _ := gorm.Open(postgres.New(postgres.Config{
					Conn: db,
				}), &gorm.Config{})

				// First query returns not found
				mock.ExpectQuery(`SELECT \* FROM "content" WHERE key = \$1 ORDER BY "content"."id" LIMIT \$2`).
					WithArgs("contact", 1).
					WillReturnRows(sqlmock.NewRows([]string{"id"}))

				// Then create new record
				mock.ExpectBegin()
				mock.ExpectQuery(`INSERT INTO "content"`).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
				mock.ExpectCommit()

				return gormDB, mock
			},
			expectedStatusCode: http.StatusOK,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var contactData ContactData
				err := json.Unmarshal(w.Body.Bytes(), &contactData)
				assert.NoError(t, err)
				assert.Equal(t, "new@example.com", contactData.Email)
				assert.Equal(t, "New York", contactData.Location)
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
			c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/contact", bytes.NewBuffer(bodyBytes))
			c.Request.Header.Set("Content-Type", "application/json")

			handler := UpdateContact(db)
			handler(c)

			assert.Equal(t, tt.expectedStatusCode, w.Code)
			if tt.validateResponse != nil {
				tt.validateResponse(t, w)
			}
		})
	}
}

// TestContentJSON tests JSON marshaling for Content
func TestContentJSON(t *testing.T) {
	content := Content{
		ID:    1,
		Key:   "test",
		Value: json.RawMessage(`{"field":"value"}`),
	}

	jsonData, err := json.Marshal(content)
	assert.NoError(t, err)

	var decoded Content
	err = json.Unmarshal(jsonData, &decoded)
	assert.NoError(t, err)

	assert.Equal(t, content.ID, decoded.ID)
	assert.Equal(t, content.Key, decoded.Key)
	assert.Equal(t, string(content.Value), string(decoded.Value))
}
