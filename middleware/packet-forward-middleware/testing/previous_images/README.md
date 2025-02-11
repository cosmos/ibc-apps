# Previous Docker Images

This section is used for upgrades e2e test.

v8.1.0

- <https://github.com/cosmos/ibc-apps/commit/04e47eb429d0cecd76d3e73a85c09d0d146aae25>
- docker build . -t pfm:v8.1.0
- docker save pfm:v8.1.0 > pfm_8_1_0.tar

If testing an upgrade, the previous version should include the corresponding
upgrade handlers, store loaders set up for the next version.
