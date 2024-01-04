# Previous Docker Images

This section is used for upgrades e2e test.

- <https://github.com/cosmos/ibc-apps/commit/2e44c421ad4a330263c58a29009b19e87562e720>
- docker build . -t icq-host:v7.1.1 -f Dockerfile.icq
- docker save icq-host:v7.1.1 > testing/previous_images/icq-host_7_1_1.tar
