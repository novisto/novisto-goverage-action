package public

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"goverage/routers/api/v1/mocks"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"goverage/data"
)

//go:generate go run -mod=mod github.com/vektra/mockery/v2 --name repository --structname Repository

func TestGetBranchBadge(t *testing.T) {
	mockRepo := mocks.NewRepository(t)
	router := NewPublicRouter(echo.New(), mockRepo)

	t.Run("ReturnsBadgeSuccessfully", func(t *testing.T) {
		mockRepo.On("GetRecentCoverage", mock.Anything, mock.Anything).Return(data.Coverage{
			RepoName:    "repo1",
			ProjectName: "project1",
			BranchName:  "branch1",
			Coverage:    90.0,
		}, nil)
		req := httptest.NewRequest(http.MethodGet, "/repos/repo1/projects/project1/branches/branch1/badge", http.NoBody)
		rec := httptest.NewRecorder()
		c := router.e.NewContext(req, rec)

		err := router.GetBranchBadge(c)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})
}
