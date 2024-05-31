package apiv1

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"

	"goverage/data"
	"goverage/internal/config"
	"goverage/internal/httperrors"

	"github.com/cohesivestack/valgo"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

type repository interface {
	ListBranches(ctx context.Context, params data.ListBranchesParams) ([]string, error)
	GetRecentCoverage(ctx context.Context, params data.GetRecentCoverageParams) (data.Coverage, error)
	GetCoverageData(ctx context.Context, params data.GetCoverageDataParams) ([]byte, error)
	ListCoverageAsc(ctx context.Context, params data.ListCoverageAscParams) ([]data.Coverage, error)
	ListCoverageDesc(ctx context.Context, params data.ListCoverageDescParams) ([]data.Coverage, error)
	UpsertCoverage(ctx context.Context, params data.UpsertCoverageParams) (data.Coverage, error)
	ListRepositories(ctx context.Context) ([]string, error)
	ListProjects(ctx context.Context, repoName string) ([]string, error)
}

type Router struct {
	e    *echo.Echo
	repo repository
}

type PythonCoverageJSONFile struct {
	Meta struct {
		Format    int    `json:"format"`
		Version   string `json:"version"`
		Timestamp string `json:"timestamp"`
	} `json:"meta"`
	Totals struct {
		CoveredLines   int     `json:"covered_lines"`
		NumStatements  int     `json:"num_statements"`
		PercentCovered float64 `json:"percent_covered"`
		MissingLines   int     `json:"missing_lines"`
	} `json:"totals"`
}

type CoverageSchema struct {
	RepoName     string    `json:"repo_name"`
	ProjectName  string    `json:"project_name"`
	BranchName   string    `json:"branch_name"`
	Commit       string    `json:"commit"`
	Coverage     float64   `json:"coverage"`
	CoverageDate time.Time `json:"coverage_date"`
}

func coverageModelToSchema(coverage data.Coverage) CoverageSchema {
	return CoverageSchema{
		RepoName:     coverage.RepoName,
		ProjectName:  coverage.ProjectName,
		BranchName:   coverage.BranchName,
		Commit:       coverage.Commit,
		Coverage:     coverage.Coverage,
		CoverageDate: coverage.CoverageDate.Time,
	}
}

func NewAPIV1Router(e *echo.Echo, repo repository) *Router {
	return &Router{e: e, repo: repo}
}

type PostCoverageRequest struct {
	RepoName    string `param:"repoName"`
	ProjectName string `param:"projectName"`
	BranchName  string `param:"branchName"`
	Commit      string `param:"commit"`
}

func (pr *PostCoverageRequest) Validate() error {
	validate := valgo.
		Is(valgo.
			String(pr.Commit, "commit").
			MinLength(8, "Commit must be at least 8 characters long"),
		)

	if !validate.Valid() {
		return validate.Error()
	}

	return nil
}

func (r *Router) PostCoverage(c echo.Context) error {
	ctx := c.Request().Context()

	var reqData PostCoverageRequest
	if err := c.Bind(&reqData); err != nil {
		return err
	}

	decodedBranchName, err := url.QueryUnescape(reqData.BranchName)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to unescape branch name")
	}
	reqData.BranchName = decodedBranchName

	if err := reqData.Validate(); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	coverageFile, err := c.FormFile("coverage")
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "coverage file is required")
	}

	src, err := coverageFile.Open()
	if err != nil {
		log.Error().Err(err).Msg("Failed to open coverage file")
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to open coverage file")
	}
	defer src.Close()

	decoder := json.NewDecoder(src)
	var coverage PythonCoverageJSONFile

	if err := decoder.Decode(&coverage); err != nil {
		// TODO: Implement a file format detection strategy to handle different coverage file formats
		return echo.NewHTTPError(http.StatusBadRequest, "failed to decode coverage file")
	}

	parsedTime, err := time.Parse("2006-01-02T15:04:05", coverage.Meta.Timestamp)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed to parse coverage timestamp")
	}

	_, err = src.Seek(0, 0)
	if err != nil {
		log.Error().Err(err).Msg("Failed to seek coverage file")
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to seek coverage file")
	}

	rawFileData, err := io.ReadAll(src)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read coverage file")
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to read coverage file")
	}

	_, err = r.repo.UpsertCoverage(ctx, data.UpsertCoverageParams{
		RepoName:    reqData.RepoName,
		ProjectName: reqData.ProjectName,
		BranchName:  reqData.BranchName,
		Commit:      reqData.Commit[:8],
		Coverage:    coverage.Totals.PercentCovered,
		CoverageDate: pgtype.Timestamptz{
			Time:  parsedTime,
			Valid: true,
		},
		RawData: rawFileData,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to upsert coverage")
		return c.String(http.StatusInternalServerError, "failed to upsert coverage")
	}

	return c.NoContent(http.StatusCreated)
}

type GetLatestBranchCoverageRequest struct {
	RepoName    string `param:"repoName"`
	ProjectName string `param:"projectName"`
	BranchName  string `param:"branchName"`
}

func (r *Router) GetLatestBranchCoverage(c echo.Context) error {
	ctx := c.Request().Context()

	var reqData GetLatestBranchCoverageRequest
	if err := c.Bind(&reqData); err != nil {
		return err
	}

	decodedBranchName, err := url.QueryUnescape(reqData.BranchName)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to unescape branch name")
	}
	reqData.BranchName = decodedBranchName

	coverage, err := r.repo.GetRecentCoverage(ctx, data.GetRecentCoverageParams{
		RepoName:    reqData.RepoName,
		ProjectName: reqData.ProjectName,
		BranchName:  reqData.BranchName,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "failed to get recent coverage")
	}

	return c.JSON(http.StatusOK, coverageModelToSchema(coverage))
}

type GetCoverageDataRequest struct {
	RepoName    string `param:"repoName"`
	ProjectName string `param:"projectName"`
	BranchName  string `param:"branchName"`
	Commit      string `param:"commit"`
}

func (r *Router) GetCoverageData(c echo.Context) error {
	ctx := c.Request().Context()

	var reqData GetCoverageDataRequest
	if err := c.Bind(&reqData); err != nil {
		return err
	}

	decodedBranchName, err := url.QueryUnescape(reqData.BranchName)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to unescape branch name")
	}
	reqData.BranchName = decodedBranchName

	coverage, err := r.repo.GetCoverageData(ctx, data.GetCoverageDataParams{
		RepoName:    reqData.RepoName,
		ProjectName: reqData.ProjectName,
		BranchName:  reqData.BranchName,
		Commit:      reqData.Commit,
	})
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "failed to get coverage data")
	}

	var jsonData map[string]interface{}
	if err := json.Unmarshal(coverage, &jsonData); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to unmarshal coverage data")
	}

	return c.JSON(http.StatusOK, jsonData)
}

type ListCoverageHistoryRequest struct {
	RepoName    string  `param:"repoName"`
	ProjectName string  `param:"projectName"`
	BranchName  string  `param:"branchName"`
	Order       *string `query:"order"`
	Limit       *int32  `query:"limit"`
	Page        *int    `query:"page"`
}

func (lr *ListCoverageHistoryRequest) SetDefaults() {
	if lr.Order == nil {
		lr.Order = lo.ToPtr("desc")
	}

	if lr.Page == nil {
		lr.Page = lo.ToPtr(1)
	}

	if lr.Limit == nil {
		lr.Limit = lo.ToPtr(int32(50))
	}
}

func (lr *ListCoverageHistoryRequest) Validate() error {
	validate := valgo.
		Is(valgo.String(*lr.Order, "order").
			InSlice([]string{"desc", "asc"}, "Order must be one of: asc, desc"),
		).
		Is(valgo.Int32(*lr.Limit, "limit").
			Between(1, 100, "Limit must be >=1 and <=100"),
		).
		Is(valgo.Int(*lr.Page, "page").
			GreaterOrEqualTo(1, "Page must be >=1"),
		)

	if !validate.Valid() {
		return validate.Error()
	}

	return nil
}

func (r *Router) ListCoverageHistory(c echo.Context) error {
	ctx := c.Request().Context()

	var reqData ListCoverageHistoryRequest
	if err := c.Bind(&reqData); err != nil {
		return err
	}

	decodedBranchName, err := url.QueryUnescape(reqData.BranchName)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to unescape branch name")
	}
	reqData.BranchName = decodedBranchName

	reqData.SetDefaults()
	if err := reqData.Validate(); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	var coverages []data.Coverage

	offset := int32((*reqData.Page - 1) * int(*reqData.Limit))

	if *reqData.Order == "asc" {
		coverages, err = r.repo.ListCoverageAsc(ctx, data.ListCoverageAscParams{
			RepoName:    reqData.RepoName,
			ProjectName: reqData.ProjectName,
			BranchName:  reqData.BranchName,
			Limit:       *reqData.Limit,
			Offset:      offset,
		})
	} else {
		coverages, err = r.repo.ListCoverageDesc(ctx, data.ListCoverageDescParams{
			RepoName:    reqData.RepoName,
			ProjectName: reqData.ProjectName,
			BranchName:  reqData.BranchName,
			Limit:       *reqData.Limit,
			Offset:      offset,
		})
	}
	if err != nil {
		log.Error().Err(err).Msg("Failed to get coverage history")
		return httperrors.WriteResponse(c, http.StatusInternalServerError, "failed to get coverage history")
	}

	coveragesSchemas := make([]CoverageSchema, 0, len(coverages))
	for _, coverage := range coverages {
		coveragesSchemas = append(coveragesSchemas, coverageModelToSchema(coverage))
	}

	return c.JSON(http.StatusOK, coveragesSchemas)
}

func (r *Router) ListRepositories(c echo.Context) error {
	ctx := c.Request().Context()

	repos, err := r.repo.ListRepositories(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get repos")
		return httperrors.WriteResponse(c, http.StatusInternalServerError, "failed to get repos")
	}
	if repos == nil {
		repos = []string{}
	}

	return c.JSON(http.StatusOK, repos)
}

type ListProjectsRequest struct {
	RepoName string `param:"repoName"`
}

func (r *Router) ListProjects(c echo.Context) error {
	ctx := c.Request().Context()

	var reqData ListProjectsRequest
	if err := c.Bind(&reqData); err != nil {
		return err
	}

	projects, err := r.repo.ListProjects(ctx, reqData.RepoName)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get projects")
		return httperrors.WriteResponse(c, http.StatusInternalServerError, "failed to get projects")
	}
	if projects == nil {
		projects = []string{}
	}

	return c.JSON(http.StatusOK, projects)
}

type ListBranchesRequest struct {
	RepoName    string `param:"repoName"`
	ProjectName string `param:"projectName"`
}

func (r *Router) ListBranches(c echo.Context) error {
	ctx := c.Request().Context()

	var reqData ListBranchesRequest
	if err := c.Bind(&reqData); err != nil {
		return err
	}

	branches, err := r.repo.ListBranches(ctx, data.ListBranchesParams{
		RepoName: reqData.RepoName, ProjectName: reqData.ProjectName,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to get branches")
		return httperrors.WriteResponse(c, http.StatusInternalServerError, "failed to get branches")
	}
	if branches == nil {
		branches = []string{}
	}

	return c.JSON(http.StatusOK, branches)
}

func (r *Router) Register() {
	apiGroup := r.e.Group("/api/v1")

	apiGroup.Use(middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
		KeyLookup: "header:X-API-Key",
		Validator: func(key string, c echo.Context) (bool, error) {
			return key == config.Config.APIKey, nil
		},
	}))

	apiGroup.GET(
		"/repos", r.ListRepositories,
	)
	apiGroup.GET(
		"/repos/:repoName/projects", r.ListProjects,
	)
	apiGroup.GET(
		"/repos/:repoName/projects/:projectName/branches", r.ListBranches,
	)
	apiGroup.POST(
		"/repos/:repoName/projects/:projectName/branches/:branchName/commits/:commit/coverage", r.PostCoverage,
	)
	apiGroup.GET(
		"/repos/:repoName/projects/:projectName/branches/:branchName/coverage", r.GetLatestBranchCoverage,
	)
	apiGroup.GET(
		"/repos/:repoName/projects/:projectName/branches/:branchName/commits/:commit/coverage_data", r.GetCoverageData,
	)
	apiGroup.GET(
		"/repos/:repoName/projects/:projectName/branches/:branchName/coverage_history", r.ListCoverageHistory,
	)
}
