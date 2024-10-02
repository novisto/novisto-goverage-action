-- name: GetRecentCoverage :one
SELECT * FROM coverage
WHERE repo_name = $1
    AND project_name = $2
    AND branch_name = $3
ORDER BY coverage_date DESC
LIMIT 1;

-- name: GetCoverageData :one
SELECT raw_data FROM coverage
WHERE repo_name = $1
    AND project_name = $2
    AND branch_name = $3
    AND "commit" = $4
LIMIT 1;

-- name: ListCoverage :many
SELECT * FROM coverage
WHERE repo_name = $1
  AND project_name = $2
  AND branch_name = $3
ORDER BY
case WHEN lower(@order_direction) = 'asc' THEN coverage_date END ASC,
case WHEN lower(@order_direction) = 'desc' THEN coverage_date END DESC,
coverage_date ASC
OFFSET $4
LIMIT $5;

-- name: UpsertCoverage :one
INSERT INTO coverage (repo_name, project_name, branch_name, commit, coverage, coverage_date, raw_data)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (repo_name, project_name, branch_name, commit)
    DO UPDATE SET coverage = $5, coverage_date = $6, raw_data = $7
RETURNING *;


-- name: ListRepositories :many
SELECT DISTINCT repo_name FROM coverage order by repo_name;

-- name: ListProjects :many
SELECT DISTINCT project_name FROM coverage WHERE repo_name = $1 order by project_name;

-- name: ListBranches :many
SELECT DISTINCT branch_name FROM coverage WHERE repo_name = $1 AND project_name = $2 order by branch_name;


-- name: ListCoverageSummary :many
SELECT repo_name, project_name, branch_name, commit, coverage, coverage_date FROM coverage
WHERE repo_name = $1
  AND project_name = $2
  AND branch_name = $3
ORDER BY
case WHEN lower(@order_direction) = 'asc' THEN coverage_date END ASC,
case WHEN lower(@order_direction) = 'desc' THEN coverage_date END DESC,
coverage_date ASC
OFFSET $4
LIMIT $5;
