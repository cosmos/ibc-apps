# Docs

Documentation for IBC apps

## Backporting to Maintained Branches

Because chains are on different versions of ibc-go, we strive to have app compatibility across older versions of ibc-go.

To do this, we maintain several branches each targeting a different version of ibc-go. You can view our maintained branches [here](https://github.com/cosmos/ibc-apps/tree/main#maintained-branches)

[`Mergify`](https://mergify.com/) has been integrated into this repo to help keep these branches in sync.

Please add the `BACKPORT` label to your PR if it should be cherry-picked into our maintained branches.

Note:

You can target any of the maintained branches. For example, if you target branch `release/v5` and add the label, the merge commit will be cherry-picked into `main` and any other maintained branch.

## Adding a new Repo

- Copy in the files from the original repo.
- Find and replace all the namespace to `github.com/cosmos/ibc-apps/modules/<MODULE_NAME>/v#` *(where # is the IBC major version)*
- Add the name to .github/labeler.yml
- Keep original proto files paths the same, unless the team wishes to move to a new namespace.
    > i.e. keep `/Stride-Labs/ibc-rate-limiting/...` instead of changing to `/cosmos/rate-limit/...` so other tools still work
    > If this is a new repo with no one using it yet in prod, you can change this without issue.
- Create that same name label in <https://github.com/cosmos/ibc-apps/labels>
- Add to the root [ReadMe](../README.md) in `List of Apps`
- Create workflow for the linting, unit testing, and e2e. The file must match the application name.