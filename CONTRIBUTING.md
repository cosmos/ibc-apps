# Contributing to ibc-apps

All development work happens directly on Github.  Both core team members and external contributers send pull requests which go through the same process.

## Repository Structure

Every app should be its own module in it's own directory.


## How can I contribute?

All contributions should abide by the [Code of Conduct](./CODE_OF_CONDUCT.md)

For issues, create a Github issue and include steps to reproduce, expected behavior, and the version affected.

For features, ensure every module has [interchain-test](https://github.com/strangelove-ventures/interchaintest) coverage.

For modules, ensure full test coverage and compatability with the main branch.


## New contributer approval process

- [ ] Reach out to community@strange.love.  
- [ ] After approval, write privileges will be granted to a member of an external team.
- [ ] Merging PRs will require approval from more than one team

Privileges will be revoked in case of failure to comply with the [Code of Conduct](../CODE_OF_CONDUCT.md)


## Versioning

We use [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

For the repository -

- the major will match the ibc-go major
- the minor/patch is dissociated from ibc-go, used for ibc-apps versioning
- the minor/patch updates in ibc-apps should update to the latest minor/patch ibc-go for the same major version.

