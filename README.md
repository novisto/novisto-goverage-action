# Novisto Goverage Action

Check coverage of Python projects, including monorepos.

## Inputs

| Name                      | Description                                                                                                                | Required | Default         |
|---------------------------|----------------------------------------------------------------------------------------------------------------------------|----------|-----------------|
| `project_name`            | Name of the project to check coverage for. This allows you to specify the name of a project in a monorepo.                 | Yes      |                 |
| `project_path`            | Path to the project to check coverage for.                                                                                 | Yes      |                 |
| `coverage_file`           | Path to the coverage file.                                                                                                 | Yes      | `coverage.json` |
| `coverage_threshold`      | Minimum coverage threshold, in %.                                                                                          | Yes      | `80`            |
| `coverage_diff_threshold` | Minimum amount, in %, that a change can drop total coverage without failing.                                               | No       | `0`             |
| `publish_coverage`        | Whether to publish coverage to Goverage. Must be `true` to enable.                                                         | Yes      | `false`         |
| `goverage_host`           | Goverage host to publish coverage to. Only required if you want to check for coverage diff or publish coverage data.       | No       |                 |
| `goverage_token`          | Goverage token to authenticate with. Only required if you want to check for coverage diff or publish coverage data.        | No       |                 |
| `github_token`            | GitHub token to authenticate with. Used for PR comments, or checking for skip comment, will use the Actions token if empty | No       |                 |

## Outputs

| Name       | Description                         |
|------------|-------------------------------------|
| `coverage` | Coverage percentage of the project. |

## Skipping coverage checks for a PR

To skip coverage checks for a PR, add a comment with the following format:

```
/goverage:skip
```


# Goverage Service

A coverage API service that allows you to publish coverage data and retrieve coverage data is provided in the `goverage` directory.

Simply deploy this service to a server, and configure the `goverage_host` and `goverage_token` inputs to use it.


# Local Development

To run the action locally, install [local-action](https://github.com/github/local-action) and 
create a `.env` file with your inputs and other environment variables.

E.g.:

```bash
INPUT_project_name=my-project
INPUT_project_path=/path/to/project
INPUT_coverage_file=coverage.json
INPUT_coverage_threshold=85
INPUT_coverage_diff_threshold=3
INPUT_publish_coverage=true
INPUT_goverage_host="http://127.0.0.1:1323"
INPUT_goverage_token=goverage-token

ACTIONS_STEP_DEBUG=true
```
