# Contributing to `ibc-apps`

All development work should happen directly on Github.  Both core team members and external contributors may send pull requests which go through the same process. Additionally:

- A release and tag for _App A_ ought not be blocked on changes in other unrelated or upstream apps
- Version management must be able to be handled independently. i.e. An _App A_ can upgrade to `ibc-go v8` and release a tag against it, while _App B_ may remain unsupported for `v8`.
- Teams with general write access to the repo should not be authorized to write to apps that they do not maintain (only default branch/tags/etc). Of course, PRs welcome :-)


## Repository Structure

Every app should be its own module in it's own directory.


## How can I contribute?

All contributions should abide by the [Code of Conduct](./CODE_OF_CONDUCT.md)

For issues, create a Github issue and include steps to reproduce, expected behavior, and the version affected.

For features, every module should have [interchain-test](https://github.com/strangelove-ventures/interchaintest) coverage.

For modules, ensure full test coverage and compatibility with the main branch.


## New contributor approval process
- [ ] Submit a Github issue titled "I should be a maintainer because..."
- [ ] After approval, write privileges will be granted to a member of an external team.
- [ ] Merging PRs will require approval from more than one team

Privileges will be revoked in case of failure to comply with the [Code of Conduct](./CODE_OF_CONDUCT.md)


## Versioning

We use [Semantic Versioning](https://semver.org/spec/v2.0.0.html) and Go modules to manage dependencies.  The main branch should build with go get.
