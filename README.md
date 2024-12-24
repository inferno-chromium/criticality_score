# Open Source Project Criticality Score (Beta)

[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/ossf/criticality_score/badge)](https://api.securityscorecards.dev/projects/github.com/ossf/criticality_score)

This project is maintained by members of the
[Securing Critical Projects WG](https://github.com/ossf/wg-securing-critical-projects).

## Goals

1. Generate a **criticality score** for every open source project.

1. Create a list of critical projects that the open source community depends on.

1. Use this data to proactively improve the security posture of these critical projects.

## Criticality Score

A project's criticality score defines the influence and importance of a project.
It is a number between
**0 (least-critical)** and **1 (most-critical)**. It is based on the following
[algorithm](https://github.com/ossf/criticality_score/blob/main/Quantifying_criticality_algorithm.pdf)
by [Rob Pike](https://github.com/robpike):

<img src="https://raw.githubusercontent.com/ossf/criticality_score/main/images/formula.png" width="359" height="96">

We use the following default parameters to derive the criticality score for an
open source project:

| Parameter (S<sub>i</sub>)  | Weight (&alpha;<sub>i</sub>) | Max threshold (T<sub>i</sub>) | Description | Reasoning |
|---|---:|---:|---|---|
| created_since | 1 | 120 | Time since the project was created (in months) | Older project has higher chance of being widely used or being dependent upon. |
| updated_since  | -1 | 120 | Time since the project was last updated (in months) | Unmaintained projects with no recent commits have higher chance of being less relied upon. |
| **contributor_count** | **2** | 5000 | Count of project contributors (with commits) | Different contributors involvement indicates project's importance. |
| org_count | 1 | 10 | Count of distinct organizations that contributors belong to | Indicates cross-organization dependency. |
| commit_frequency | 1 | 1000 | Average number of commits per week in the last year | Higher code churn has slight indication of project's importance. Also, higher susceptibility to vulnerabilities.
| recent_releases_count | 0.5 | 26 | Number of releases in the last year | Frequent releases indicates user dependency. Lower weight since this is not always used. |
| closed_issues_count | 0.5 | 5000 | Number of issues closed in the last 90 days | Indicates high contributor involvement and focus on closing user issues. Lower weight since it is dependent on project contributors. |
| updated_issues_count | 0.5 | 5000 | Number of issues updated in the last 90 days | Indicates high contributor involvement. Lower weight since it is dependent on project contributors. |
| comment_frequency | 1 | 15 | Average number of comments per issue in the last 90 days | Indicates high user activity and dependence. |
| **dependents_count** | **2** | 500000 | Number of project mentions in the commit messages | Indicates repository use, usually in version rolls. This parameter works across all languages, including C/C++ that don't have package dependency graphs (though hack-ish). Plan to add package dependency trees in the near future. |

**NOTE**:

- You can override those default values at runtime as described below.
- We are looking for community ideas to improve upon these parameters.
- There will always be exceptions to the individual reasoning rules.

## Usage

```shell
$ go install github.com/ossf/criticality_score/v2/cmd/criticality_score@latest

$ export GITHUB_TOKEN=...         # requires a GitHub token to work
$ gcloud auth login --update-adc  # optional, add -depsdev-disable to skip

$ criticality_score -gcp-project-id=[your projectID] https://github.com/kubernetes/kubernetes
repo.name: kubernetes
repo.url: https://github.com/kubernetes/kubernetes
repo.language: Go
repo.license: Apache License 2.0
legacy.created_since: 87
legacy.updated_since: 0
legacy.contributor_count: 3999
legacy.watchers_count: 79583
legacy.org_count: 5
legacy.commit_frequency: 97.2
legacy.recent_releases_count: 70
legacy.updated_issues_count: 5395
legacy.closed_issues_count: 3062
legacy.comment_frequency: 5.5
legacy.dependents_count: 454393
default_score: 0.99107
```

The score can be changed by using the `-scoring-config` parameter and supplying
a different configuration file to specify how the score is calculated.

By default the `original_pike.yml` configuration is used to calculate the score.
However, other config files can be supplied to produce different scores. See
[config/scorer](`https://github.com/ossf/criticality_score/blob/main/config/scorer`) for more.

Feel free to copy one of the configurations and adjust the weights and
thresholds to suit your needs.

### Authentication

Before running criticality score, you need to:

- For GitHub repos, you need to
[create a GitHub access token](https://docs.github.com/en/free-pro-team@latest/developers/apps/about-apps#personal-access-tokens)
and set it in environment variable `GITHUB_AUTH_TOKEN`.
This helps to avoid the GitHub's
[api rate limits](https://developer.github.com/v3/#rate-limiting)
with unauthenticated requests.

```shell
# For posix platforms, e.g. linux, mac:
export GITHUB_AUTH_TOKEN=<your access token>

# For windows:
set GITHUB_AUTH_TOKEN=<your access token>
```

<!-- Hide GitLab documentation until support is added back. -->
<!--
- For GitLab repos, you need to
[create a GitLab access token](https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html)
and set it in environment variable `GITLAB_AUTH_TOKEN`.
This helps to avoid the GitLab's api limitations for unauthenticated users.

```shell
# For posix platforms, e.g. linux, mac:
export GITLAB_AUTH_TOKEN=<your access token>

# For windows:
set GITLAB_AUTH_TOKEN=<your access token>
```
-->

### Formatting Results

There are three formats currently: `text`, `json`, and `csv`. Others may be added in the future.

These may be specified with the `-format` flag.

## Other Commands

The criticality score project also has other commands for generating and
working with criticality score data.

- [`enumerate_github`](https://github.com/ossf/criticality_score/blob/main/cmd/enumerate_github):
  a tool for accurately collecting a set of GitHub repos with a minimum number of stars
- [`collect_signals`](https://github.com/ossf/criticality_score/blob/main/cmd/collect_signals):
  a worker for collecting raw signals at scale by leveraging the
  [Scorecard project's](https://github.com/ossf/scorecard) infrastructure.
- [`scorer`](https://github.com/ossf/criticality_score/blob/main/cmd/scorer):
  a tool for recalculating criticality scores based on an input CSV file.

## Public Data

If you're interested in seeing a list of critical projects with their criticality
score, we publish them in `csv` format and a BigQuery dataset.

This data is generated using a production instance of the criticality score
project running in GCP. Details for how this is deployed can be found in the
[infra](https://github.com/ossf/criticality_score/blob/main/infra) directory.

**NOTE**: Currently, these lists are derived from **projects hosted on GitHub ONLY**.
We do plan to expand them in near future to account for projects hosted on other
source control systems.

### CSV data

The data is available on Google Cloud Storage and can be downloaded via:

- web browser: [commondatastorage.googleapis.com/ossf-criticality-score/index.html?prefix=2022.11.28/](https://commondatastorage.googleapis.com/ossf-criticality-score/index.html?prefix=2022.11.28/)
- [`gsutil`](https://cloud.google.com/storage/docs/gsutil_install)
command-line tool: `gsutil ls gs://ossf-criticality-score/`

### BigQuery Dataset

This data is available in the public [BigQuery dataset](https://console.cloud.google.com/bigquery?d=criticality_score_cron&p=openssf&t=criticality-score-v0-latest&page=table).

With a GCP account you can run queries across the data. For example, here is a query returning the top 100 repos by score:

```sql
  SELECT repo.url, default_score
    FROM `openssf.criticality_score_cron.criticality-score-v0-latest`
ORDER BY default_score DESC
   LIMIT 100;
```

## Contributing

If you want to get involved or have ideas you'd like to chat about, we discuss this project in the [Securing Critical Projects WG](https://github.com/ossf/wg-securing-critical-projects) meetings.

See the [Community Calendar](https://calendar.google.com/calendar?cid=czYzdm9lZmhwNWk5cGZsdGI1cTY3bmdwZXNAZ3JvdXAuY2FsZW5kYXIuZ29vZ2xlLmNvbQ) for the schedule and meeting invitations.

See the [Contributing](CONTRIBUTING.md) documentation for guidance on how to contribute.
