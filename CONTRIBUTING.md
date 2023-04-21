# Contributing to folderr-cli

Hi! We're glad you want to contribute!

We have a few guidelines first

Note: These guidelines are applicable to all projects owned by the Folderr organization written in go.

Keep NSFW out of the commits and PR, thanks!

## New contributor guide

Please read the [README](./README.md) first.

Here are some requirements:
- Ensure you're using (go)[https://go.dev] 1.20.2 or later

## Creating a good issue:

Creating issues is important, and creating good issues can be hard.
For this we have a simple set of guidelines.

Please [see if the issue is already present!](https://docs.github.com/en/github/searching-for-information-on-github/searching-on-github/searching-issues-and-pull-requests#search-by-the-title-body-or-comments)


Then create an issue with the following information:

- What is the issue?
- State your OS and its version
- State the version of the application your using 
  - clis will have their version from `<command> -version`
- If you're building the app, state the version of Go you're using.
  - If its earlier than 1.20.2 please update your go version!
- Provide any and all extra detail
- If any debug logs are generated, please include those!
  - Redact ANY private information (access keys, urls to personal images, emails, etc)
- Reproduction steps. An issue is useless without these!
  - How do we replicate the issue? does it happen with specific input?

## Creating a good PR

- Use an appropriate title
- What does the PR solve?
- Link to any issues if the PR solves them
- Vet and lint your code.
  - Lint with staticcheck
- Ensure the PR is using at least go 1.20.2 or later
- Use descriptive commit messages!
- If you solved an issue, please link your [PR to the issue](https://docs.github.com/en/issues/tracking-your-work-with-issues/linking-a-pull-request-to-an-issue).
- Allow [maintainer edits](https://docs.github.com/en/github/collaborating-with-issues-and-pull-requests/allowing-changes-to-a-pull-request-branch-created-from-a-fork) so we can update for a merge.
- We may ask for changes before submitting your pr either via [suggested changes](https://docs.github.com/en/github/collaborating-with-issues-and-pull-requests/incorporating-feedback-in-your-pull-request) or pull request comments
- Please [resolve](https://docs.github.com/en/github/collaborating-with-issues-and-pull-requests/commenting-on-a-pull-request#resolving-conversations) conversations as you update your PR

Git issues? Follow this [git tutorial](https://github.com/skills/resolve-merge-conflicts)

## Commits

- Ensure you use `go vet` and `staticcheck` before committing!
- Please test first. (`go test ./cmd` for folderr-cli)
- Use descriptive commit messages
- Sign your commits! We don't publish unsigned or unverified commits.

## What is a good first PR?

To find a good PR, please visit our [issues page](https://github.com/Folderr/folderr-cli/issues) and find an issue you believe you can fix. They are labelled to help find issues!

After fixing the issue please follow both the [commit](#commits) and [pull request](#creating-a-good-pr) guidelines

Issues will be labelled for help identifying severity, and what needs to be fixed, as well as what the issue *is*


## My PR is merged. Now what?

Congragulations!
Your efforts are now visible in our repository and will likely be included in the next release!

If you want to continue contributing, feel free!