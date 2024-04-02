const core = require('@actions/core');
const github = require('@actions/github');

const readFile = require('fs/promises');
const dedent = require('dedent-js');

const REF_TAGS_PREFIX = "refs/tags/"
const REF_HEADS_PREFIX = "refs/heads/"

const SUCCESS_EMOJI = '✅';
const FAILURE_EMOJI = ':x:';
const WARNING_EMOJI = ':warning:';

const SKIP_COVERAGE_COMMENT = `/goverage:skip`;


function round(number, decimals) {
    if (decimals === undefined) {
        decimals = 2;
    }
    return Math.round(number * Math.pow(10, decimals)) / Math.pow(10, decimals);
}

async function checkForSkipComment(githubToken) {
    // Check if there's a skip comment on the PR
    const owner = process.env.OVERRIDE_REPO_OWNER || github.context.repo.owner;
    const repo = process.env.OVERRIDE_REPO_NAME || github.context.repo.repo;
    const number = process.env.OVERRIDE_PR_NUMBER || github.context.payload.pull_request.number;

    const octokit = github.getOctokit(githubToken);
    const comments = await octokit.rest.issues.listComments({
        owner: owner,
        repo: repo,
        issue_number: number
    });

    const skipComment = comments.data.find(comment => {
        return comment.body.includes(SKIP_COVERAGE_COMMENT);
    });

    return skipComment !== undefined;
}

function getPrefixString(projectName) {
    return `:gear: Coverage for **${projectName}**`;
}

function formatComment(inputs, currentCoverage, files, diffCheck) {
    const emoji = currentCoverage < inputs.coverageThreshold ? FAILURE_EMOJI : SUCCESS_EMOJI;
    const adjective = currentCoverage < inputs.coverageThreshold ? 'below' : 'above';

    let diffThresholdMessageString = ''
    if (diffCheck.checked) {
        if (diffCheck.success) {
            if (diffCheck.change < 0) {
                diffThresholdMessageString = `${WARNING_EMOJI} Coverage decreased by **${Math.abs(round(diffCheck.change, 2))}%** compared to base branch **${diffCheck.baseRef}**.`;
            } else {
                diffThresholdMessageString = `${SUCCESS_EMOJI} Coverage increased by **${round(diffCheck.change, 2)}%** compared to base branch **${diffCheck.baseRef}**.`;
            }
        } else {
            diffThresholdMessageString = `${FAILURE_EMOJI} Coverage decreased by **${round(diffCheck.change, 2)}%** compared to base branch **${diffCheck.baseRef}**, more than the specified threshold of **${inputs.coverageDiffThreshold}%**.`;
        }
    }

    let modulesTable = '';
    files.forEach((elem) => {
        modulesTable += `| \`${elem[0]}\` | ${elem[1].join(", ")} | ${round(elem[2])}% |\n`
    });

    return dedent(`${getPrefixString(inputs.projectName)}
    ---
    ${emoji} Coverage of **${round(currentCoverage)}%** is ${adjective} threshold of **${inputs.coverageThreshold}%**.
    ${diffThresholdMessageString}
    
    <details>
    <summary>Missing Coverage Details</summary>
    <br>
    
    | File | Missing Lines | Coverage |
    | --- | --- | --- |
    ${modulesTable}
    
    </details>
    `);
}

async function postCommentOnPR(inputs, coverageJSON, diffCheck) {
    let files = Object.entries(coverageJSON.files).reduce((acc, [key, value]) => {
        if (value.summary.missing_lines > 0) {
            acc.push([key, value.missing_lines, value.summary.percent_covered]);
        }
        return acc;
    }, []);

    const comment = formatComment(
        inputs,
        coverageJSON.totals.percent_covered,
        files,
        diffCheck,
    );
    core.debug(`\n${comment}`);

    const octokit = github.getOctokit(inputs.githubToken);

    const owner = process.env.OVERRIDE_REPO_OWNER || github.context.repo.owner;
    const repo = process.env.OVERRIDE_REPO_NAME || github.context.repo.repo;
    const number = process.env.OVERRIDE_PR_NUMBER || github.context.payload.pull_request.number;

    core.info(`Fetching comments for PR #${number} on ${owner}/${repo}`);

    const comments = await octokit.rest.issues.listComments({
        owner: owner,
        repo: repo,
        issue_number: number
    });

    const matchingComment = comments.data.find(comment => {
        return comment.body.includes(getPrefixString(inputs.projectName));
    });

    if (matchingComment !== undefined) {
        core.info(`Updating comment ${matchingComment.id} with new coverage details`);
        await octokit.rest.issues.updateComment({
            owner: owner,
            repo: repo,
            comment_id: matchingComment.id,
            body: comment
        });
    } else {
        core.info('Creating comment with coverage details');
        await octokit.rest.issues.createComment({
            owner: owner,
            repo: repo,
            issue_number: number,
            body: comment
        });
    }
}

async function compareCoverageWithBaseRef(inputs, currentCoverage) {
    const repo = process.env.OVERRIDE_REPO_NAME || github.context.repo.repo;

    const baseRef = process.env.OVERRIDE_BASE_REF || github.context.payload.pull_request.base.ref;
    const urlEncodedBaseRef = encodeURIComponent(baseRef);

    // call goverage to get the current coverage
    core.info(`Fetching coverage for ${inputs.projectName} on ${baseRef} from ${inputs.goverageHost}`);

    const response = await fetch(
        `${inputs.goverageHost}/api/v1/repos/${repo}/projects/${inputs.projectName}/branches/${urlEncodedBaseRef}/coverage`, {
            headers: {
                "X-API-Key": inputs.goverageToken
            },
        }
    );

    let checked = false;
    let success = false;
    let change = 0;

    if (response.status === 404) {
        core.warning(`Coverage not found for ${inputs.projectName} on ${baseRef} from ${inputs.goverageHost}`);
    } else if (response.status !== 200) {
        core.setFailed(`Failed to fetch coverage from ${inputs.goverageHost}`);
    } else {
        checked = true;

        const data = await response.json();

        const baseCoverage = data.coverage;
        change = currentCoverage - baseCoverage;

        if (currentCoverage < baseCoverage - inputs.coverageDiffThreshold) {
            core.setFailed(`Coverage decreased by ${round(change, 2)}% compared to ${baseRef} which is more than the allowed threshold of ${inputs.coverageDiffThreshold}%`);
            success = false;
        } else {
            core.info(`Coverage changed by ${round(change, 2)}% compared to ${baseRef}`);
            success = true;
        }
    }

    return {
        checked: checked,
        success: success,
        change: change,
        baseRef: baseRef,
    }
}

async function run() {
    try {
        core.info('Reading inputs');
        const inputs = {
            projectName: core.getInput('project_name', {required: true}),
            projectPath: core.getInput('project_path', {required: true}),
            coverageFile: core.getInput('coverage_file', {required: true}),
            coverageThreshold: parseInt(core.getInput('coverage_threshold', {required: true})),
            coverageDiffThreshold: parseInt(core.getInput('coverage_diff_threshold', {required: false})),
            publishCoverage: core.getInput('publish_coverage', {required: true}) === 'true',
            goverageHost: core.getInput('goverage_host', {required: false}),
            goverageToken: core.getInput('goverage_token', {required: false}),
            githubToken: core.getInput('github_token', {required: false}) || process.env.GITHUB_TOKEN,
        }

        // Validation
        if (inputs.coverageThreshold < 0 || inputs.coverageThreshold > 100) {
            throw new Error('coverage_threshold must be between 0 and 100');
        }

        if (inputs.coverageDiffThreshold < 0 || inputs.coverageDiffThreshold > 100) {
            throw new Error('coverage_diff_threshold must be between 0 and 100');
        }

        if (!inputs.githubToken) {
            core.warning('github_token not provided, skipping comments');
        }

        if (inputs.githubToken && github.context.eventName === 'pull_request') {
            core.info('Checking for skip comment on PR')
            const skipComment = await checkForSkipComment(inputs.githubToken);
            if (skipComment) {
                core.warning('Skip comment found, skipping coverage check');
                return;
            }
        }

        core.info(`Reading coverage file: ${inputs.projectPath}/${inputs.coverageFile}`);
        const coverage = await readFile(`${inputs.projectPath}/${inputs.coverageFile}`, 'utf8');
        const coverageJSON = JSON.parse(coverage);

        core.info('Checking coverage threshold');
        const currentCoverage = coverageJSON.totals.percent_covered;
        core.setOutput("coverage", currentCoverage);

        if (currentCoverage < inputs.coverageThreshold) {
            const message = `Coverage is below threshold: ${currentCoverage}%`;
            core.setFailed(message);
        } else {
            const message = `Coverage is above threshold: ${currentCoverage}%`;
            core.info(message);
        }

        if (github.context.eventName === 'pull_request') {
            let diffCheck = {};

            if (inputs.coverageDiffThreshold !== 0 && inputs.goverageHost !== "" && inputs.goverageToken !== "") {
                diffCheck = await compareCoverageWithBaseRef(inputs, currentCoverage);
            }

            if (inputs.githubToken) {
                await postCommentOnPR(inputs, coverageJSON, diffCheck);
            }
        }

        if (inputs.publishCoverage && inputs.goverageHost !== "" && inputs.goverageToken !== "") {
            core.info('Publishing coverage');
            const repo = process.env.OVERRIDE_REPO_NAME || github.context.repo.repo;
            const commitHash = process.env.OVERRIDE_COMMIT_HASH || github.context.sha;

            const ref = process.env.OVERRIDE_REF || github.context.ref;
            let refName;
            if (ref.startsWith(REF_TAGS_PREFIX)) {
                refName = ref.substring(REF_HEADS_PREFIX.length)
            } else if (ref.startsWith(REF_HEADS_PREFIX)) {
                refName = ref.substring(REF_HEADS_PREFIX.length)
            } else {
                refName = ref;
            }

            const formData = new FormData();
            formData.append("coverage", new Blob([coverage]), "coverage.json");

            const response = await fetch(
                `${inputs.goverageHost}/api/v1/repos/${repo}/projects/${inputs.projectName}/branches/${refName}/commits/${commitHash}/coverage`, {
                    method: 'POST',
                    headers: {
                        "X-API-Key": inputs.goverageToken
                    },
                    body: formData,
                }
            )

            if (response.status !== 201) {
                core.warning(`Failed to publish coverage to ${inputs.goverageHost}`);
            }
        }
    } catch (error) {
        core.setFailed(error.message);
    }
}

module.exports = { run }
