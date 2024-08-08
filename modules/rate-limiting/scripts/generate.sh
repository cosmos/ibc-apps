cd proto
buf generate --template buf.gen.gogo.yaml
buf generate --template buf.gen.pulsar.yaml
cd ..

cp -r github.com/cosmos/ibc-apps/modules/rate-limiting/v8/* ./
rm -rf github.com
rm -rf ratelimit
