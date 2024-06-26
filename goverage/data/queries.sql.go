// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0
// source: queries.sql

package data

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const getCoverageData = `-- name: GetCoverageData :one
SELECT raw_data FROM coverage
WHERE repo_name = $1
    AND project_name = $2
    AND branch_name = $3
    AND "commit" = $4
LIMIT 1
`

type GetCoverageDataParams struct {
	RepoName    string
	ProjectName string
	BranchName  string
	Commit      string
}

func (q *Queries) GetCoverageData(ctx context.Context, arg GetCoverageDataParams) ([]byte, error) {
	row := q.db.QueryRow(ctx, getCoverageData,
		arg.RepoName,
		arg.ProjectName,
		arg.BranchName,
		arg.Commit,
	)
	var raw_data []byte
	err := row.Scan(&raw_data)
	return raw_data, err
}

const getRecentCoverage = `-- name: GetRecentCoverage :one
SELECT id, repo_name, project_name, branch_name, commit, coverage, coverage_date, raw_data FROM coverage
WHERE repo_name = $1
    AND project_name = $2
    AND branch_name = $3
ORDER BY coverage_date DESC
LIMIT 1
`

type GetRecentCoverageParams struct {
	RepoName    string
	ProjectName string
	BranchName  string
}

func (q *Queries) GetRecentCoverage(ctx context.Context, arg GetRecentCoverageParams) (Coverage, error) {
	row := q.db.QueryRow(ctx, getRecentCoverage, arg.RepoName, arg.ProjectName, arg.BranchName)
	var i Coverage
	err := row.Scan(
		&i.ID,
		&i.RepoName,
		&i.ProjectName,
		&i.BranchName,
		&i.Commit,
		&i.Coverage,
		&i.CoverageDate,
		&i.RawData,
	)
	return i, err
}

const listBranches = `-- name: ListBranches :many
SELECT DISTINCT branch_name FROM coverage WHERE repo_name = $1 AND project_name = $2 order by branch_name
`

type ListBranchesParams struct {
	RepoName    string
	ProjectName string
}

func (q *Queries) ListBranches(ctx context.Context, arg ListBranchesParams) ([]string, error) {
	rows, err := q.db.Query(ctx, listBranches, arg.RepoName, arg.ProjectName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []string
	for rows.Next() {
		var branch_name string
		if err := rows.Scan(&branch_name); err != nil {
			return nil, err
		}
		items = append(items, branch_name)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listCoverageAsc = `-- name: ListCoverageAsc :many
SELECT id, repo_name, project_name, branch_name, commit, coverage, coverage_date, raw_data FROM coverage
WHERE repo_name = $1
  AND project_name = $2
  AND branch_name = $3
ORDER BY coverage_date ASC
OFFSET $4
LIMIT $5
`

type ListCoverageAscParams struct {
	RepoName    string
	ProjectName string
	BranchName  string
	Offset      int32
	Limit       int32
}

func (q *Queries) ListCoverageAsc(ctx context.Context, arg ListCoverageAscParams) ([]Coverage, error) {
	rows, err := q.db.Query(ctx, listCoverageAsc,
		arg.RepoName,
		arg.ProjectName,
		arg.BranchName,
		arg.Offset,
		arg.Limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Coverage
	for rows.Next() {
		var i Coverage
		if err := rows.Scan(
			&i.ID,
			&i.RepoName,
			&i.ProjectName,
			&i.BranchName,
			&i.Commit,
			&i.Coverage,
			&i.CoverageDate,
			&i.RawData,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listCoverageDesc = `-- name: ListCoverageDesc :many
SELECT id, repo_name, project_name, branch_name, commit, coverage, coverage_date, raw_data FROM coverage
WHERE repo_name = $1
    AND project_name = $2
    AND branch_name = $3
ORDER BY coverage_date DESC
OFFSET $4
LIMIT $5
`

type ListCoverageDescParams struct {
	RepoName    string
	ProjectName string
	BranchName  string
	Offset      int32
	Limit       int32
}

func (q *Queries) ListCoverageDesc(ctx context.Context, arg ListCoverageDescParams) ([]Coverage, error) {
	rows, err := q.db.Query(ctx, listCoverageDesc,
		arg.RepoName,
		arg.ProjectName,
		arg.BranchName,
		arg.Offset,
		arg.Limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Coverage
	for rows.Next() {
		var i Coverage
		if err := rows.Scan(
			&i.ID,
			&i.RepoName,
			&i.ProjectName,
			&i.BranchName,
			&i.Commit,
			&i.Coverage,
			&i.CoverageDate,
			&i.RawData,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listProjects = `-- name: ListProjects :many
SELECT DISTINCT project_name FROM coverage WHERE repo_name = $1 order by project_name
`

func (q *Queries) ListProjects(ctx context.Context, repoName string) ([]string, error) {
	rows, err := q.db.Query(ctx, listProjects, repoName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []string
	for rows.Next() {
		var project_name string
		if err := rows.Scan(&project_name); err != nil {
			return nil, err
		}
		items = append(items, project_name)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const listRepositories = `-- name: ListRepositories :many
SELECT DISTINCT repo_name FROM coverage order by repo_name
`

func (q *Queries) ListRepositories(ctx context.Context) ([]string, error) {
	rows, err := q.db.Query(ctx, listRepositories)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []string
	for rows.Next() {
		var repo_name string
		if err := rows.Scan(&repo_name); err != nil {
			return nil, err
		}
		items = append(items, repo_name)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const upsertCoverage = `-- name: UpsertCoverage :one
INSERT INTO coverage (repo_name, project_name, branch_name, commit, coverage, coverage_date, raw_data)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (repo_name, project_name, branch_name, commit)
    DO UPDATE SET coverage = $5, coverage_date = $6, raw_data = $7
RETURNING id, repo_name, project_name, branch_name, commit, coverage, coverage_date, raw_data
`

type UpsertCoverageParams struct {
	RepoName     string
	ProjectName  string
	BranchName   string
	Commit       string
	Coverage     float64
	CoverageDate pgtype.Timestamptz
	RawData      []byte
}

func (q *Queries) UpsertCoverage(ctx context.Context, arg UpsertCoverageParams) (Coverage, error) {
	row := q.db.QueryRow(ctx, upsertCoverage,
		arg.RepoName,
		arg.ProjectName,
		arg.BranchName,
		arg.Commit,
		arg.Coverage,
		arg.CoverageDate,
		arg.RawData,
	)
	var i Coverage
	err := row.Scan(
		&i.ID,
		&i.RepoName,
		&i.ProjectName,
		&i.BranchName,
		&i.Commit,
		&i.Coverage,
		&i.CoverageDate,
		&i.RawData,
	)
	return i, err
}
