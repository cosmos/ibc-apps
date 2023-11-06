# Previous Docker Images

This section is used for upgrades e2e test.

v7.0.0

- <https://github.com/cosmos/ibc-apps/commit/e5a274cf6fc2eb965a9f8da4bdeb7c718d06661d>
- docker build . -t icq-host:v7.0.0 -f Dockerfile.icq
- docker save icq-host:v7.0.0 > icq-host_7_0_0.tar
