# Docs

Documentation for IBC apps

## Backporting to Maintained Branches

Because chains are on different versions of ibc-go, we strive to have app compatibility across older versions of ibc-go.

To do this, we maintain several branches each targeting a different version of ibc-go. You can view our maintained branches [here](https://github.com/cosmos/ibc-apps/tree/main#maintained-branches)


[`Mergify`](https://mergify.com/) has been integrated into this repo to help keep these branches in sync.


Please add the `BACKPORT` label to your PR if it should be cherry-picked into our maintained branches.


Note:

You can target any of the maintained branches. For example, if you target branch `release/v5` and add the label, the merge commit will be cherry-picked into `main` and any other maintained branch.

