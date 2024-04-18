package apiv1

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"goverage/data"
	"goverage/routers/api/v1/mocks"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

//go:generate go run -mod=mod github.com/vektra/mockery/v2 --name repository --structname Repository

func TestListRepository(t *testing.T) {
	setup := func(returnValue interface{}, returnError error) (*Router, echo.Context, *httptest.ResponseRecorder) {
		mockDB := new(mocks.Repository)
		router := NewAPIV1Router(echo.New(), mockDB)
		mockDB.On("ListRepositories", mock.Anything).Return(returnValue, returnError)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/repos", http.NoBody)
		req.ContentLength = 0 // Required for echo to parse the request body correctly
		rec := httptest.NewRecorder()
		c := router.e.NewContext(req, rec)

		return router, c, rec
	}

	t.Run("ReturnsReposSuccessfully", func(t *testing.T) {
		router, c, rec := setup([]string{"repo1", "repo2"}, nil)

		err := router.ListRepositories(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, `["repo1","repo2"]`, strings.Trim(rec.Body.String(), "\n"))
	})

	t.Run("ReturnsEmptyWhenNoRepos", func(t *testing.T) {
		router, c, rec := setup(nil, nil)

		err := router.ListRepositories(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, `[]`, strings.Trim(rec.Body.String(), "\n"))
	})

	t.Run("ReturnsErrorWhenDBFails", func(t *testing.T) {
		router, c, rec := setup(nil, errors.New("db error"))

		err := router.ListRepositories(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestListProjects(t *testing.T) {
	expectedProjectsParam := "repo1"

	setup := func(returnValue interface{}, returnError error) (*Router, echo.Context, *httptest.ResponseRecorder) {
		mockDB := new(mocks.Repository)
		router := NewAPIV1Router(echo.New(), mockDB)
		mockDB.On("ListProjects", mock.Anything, expectedProjectsParam).Return(returnValue, returnError)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/repos/repo1/projects", http.NoBody)
		req.ContentLength = 0 // Required for echo to parse the request body correctly
		rec := httptest.NewRecorder()
		c := router.e.NewContext(req, rec)
		c.SetParamNames("repoName")
		c.SetParamValues(expectedProjectsParam)

		return router, c, rec
	}

	t.Run("ReturnsProjectsSuccessfully", func(t *testing.T) {
		router, c, rec := setup([]string{"project1", "project2"}, nil)

		err := router.ListProjects(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, `["project1","project2"]`, strings.Trim(rec.Body.String(), "\n"))
	})

	t.Run("ReturnsEmptyWhenNoProjects", func(t *testing.T) {
		router, c, rec := setup(nil, nil)

		err := router.ListProjects(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, `[]`, strings.Trim(rec.Body.String(), "\n"))
	})

	t.Run("ReturnsErrorWhenDBFails", func(t *testing.T) {
		router, c, rec := setup(nil, errors.New("db error"))

		err := router.ListProjects(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestListBranches(t *testing.T) {
	expectedBranchesParams := data.ListBranchesParams{
		RepoName:    "repo1",
		ProjectName: "project1",
	}

	setup := func(returnValue interface{}, returnError error) (*Router, echo.Context, *httptest.ResponseRecorder) {
		mockDB := new(mocks.Repository)
		router := NewAPIV1Router(echo.New(), mockDB)
		mockDB.On("ListBranches", mock.Anything, expectedBranchesParams).Return(returnValue, returnError)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/repos/repo1/projects/project1/branches", http.NoBody)
		req.ContentLength = 0 // Required for echo to parse the request body correctly
		rec := httptest.NewRecorder()
		c := router.e.NewContext(req, rec)
		c.SetParamNames("repoName", "projectName")
		c.SetParamValues(expectedBranchesParams.RepoName, expectedBranchesParams.ProjectName)

		return router, c, rec
	}

	t.Run("ReturnsBranchesSuccessfully", func(t *testing.T) {
		router, c, rec := setup([]string{"branch1", "branch2"}, nil)

		err := router.ListBranches(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, `["branch1","branch2"]`, strings.Trim(rec.Body.String(), "\n"))
	})

	t.Run("ReturnsEmptyWhenNoBranches", func(t *testing.T) {
		router, c, rec := setup(nil, nil)

		err := router.ListBranches(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, `[]`, strings.Trim(rec.Body.String(), "\n"))
	})

	t.Run("ReturnsErrorWhenDBFails", func(t *testing.T) {
		router, c, rec := setup(nil, errors.New("db error"))

		err := router.ListBranches(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}
