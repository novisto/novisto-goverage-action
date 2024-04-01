-- +goose Up
-- +goose StatementBegin
CREATE TABLE coverage (
    id SERIAL PRIMARY KEY,
    repo_name VARCHAR(255) NOT NULL,
    project_name VARCHAR(255) NOT NULL,
    branch_name VARCHAR(255) NOT NULL,
    commit VARCHAR(255) NOT NULL,
    coverage FLOAT NOT NULL,
    coverage_date TIMESTAMPTZ NOT NULL,
    raw_data JSONB NOT NULL
);

CREATE INDEX coverage_repo_name_idx ON coverage (repo_name);
CREATE INDEX coverage_project_name_idx ON coverage (project_name);
CREATE INDEX coverage_branch_name_idx ON coverage (branch_name);
CREATE INDEX coverage_date_idx ON coverage (coverage_date);

CREATE UNIQUE INDEX coverage_repo_name_project_name_branch_name_commit_idx
    ON coverage (repo_name, project_name, branch_name, commit);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE coverage;
-- +goose StatementEnd
