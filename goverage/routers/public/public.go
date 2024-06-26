package public

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strings"

	"goverage/data"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

type Router struct {
	e    *echo.Echo
	repo repository
}

type repository interface {
	GetRecentCoverage(ctx context.Context, params data.GetRecentCoverageParams) (data.Coverage, error)
}

func NewPublicRouter(e *echo.Echo, repo repository) *Router {
	return &Router{e: e, repo: repo}
}

type GetBranchBadgeRequest struct {
	RepoName    string `param:"repoName"`
	ProjectName string `param:"projectName"`
	BranchName  string `param:"branchName"`
}

func (r *Router) GetBranchBadge(c echo.Context) error {
	ctx := context.Background()

	var reqData GetBranchBadgeRequest
	if err := c.Bind(&reqData); err != nil {
		return err
	}

	dbCoverage, err := r.repo.GetRecentCoverage(ctx, data.GetRecentCoverageParams{
		RepoName:    reqData.RepoName,
		ProjectName: reqData.ProjectName,
		BranchName:  reqData.BranchName,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	url := fmt.Sprintf(
		"https://img.shields.io/badge/%s%%2F%s_%s-%s%%25-%s",
		strings.ReplaceAll(dbCoverage.RepoName, "-", "--"),
		strings.ReplaceAll(dbCoverage.ProjectName, "-", "--"),
		strings.ReplaceAll(dbCoverage.BranchName, "-", "--"),
		fmt.Sprintf("%.0f", math.Round(dbCoverage.Coverage)),
		"blue",
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create badge request")
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to fetch badge")
	}

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("Failed to fetch badge")
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to fetch badge")
	}
	defer resp.Body.Close()

	return c.Stream(http.StatusOK, "image/svg+xml", resp.Body)
}

func (r *Router) Register() {
	r.e.GET("/repos/:repoName/projects/:projectName/branches/:branchName/badge", r.GetBranchBadge)
}
