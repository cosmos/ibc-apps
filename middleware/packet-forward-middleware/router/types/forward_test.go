package types_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v4/router/types"
	"github.com/stretchr/testify/require"
)

func TestForwardMetadataUnmarshalStringNext(t *testing.T) {
	const memo = "{\"forward\":{\"receiver\":\"noble1f4cur2krsua2th9kkp7n0zje4stea4p9tu70u8\",\"port\":\"transfer\",\"channel\":\"channel-0\",\"timeout\":0,\"next\":\"{\\\"forward\\\":{\\\"receiver\\\":\\\"noble1l505zhahp24v5jsmps9vs5asah759fdce06sfp\\\",\\\"port\\\":\\\"transfer\\\",\\\"channel\\\":\\\"channel-0\\\",\\\"timeout\\\":0}}\"}}"
	var packetMetadata types.PacketMetadata

	err := json.Unmarshal([]byte(memo), &packetMetadata)
	require.NoError(t, err)

	nextBz, err := json.Marshal(packetMetadata.Forward.Next)
	require.NoError(t, err)
	require.Equal(t, `{"forward":{"receiver":"noble1l505zhahp24v5jsmps9vs5asah759fdce06sfp","port":"transfer","channel":"channel-0","timeout":0}}`, string(nextBz))
}

func TestForwardMetadataUnmarshalJSONNext(t *testing.T) {
	const memo = "{\"forward\":{\"receiver\":\"noble1f4cur2krsua2th9kkp7n0zje4stea4p9tu70u8\",\"port\":\"transfer\",\"channel\":\"channel-0\",\"timeout\":0,\"next\":{\"forward\":{\"receiver\":\"noble1l505zhahp24v5jsmps9vs5asah759fdce06sfp\",\"port\":\"transfer\",\"channel\":\"channel-0\",\"timeout\":0}}}}"
	var packetMetadata types.PacketMetadata

	err := json.Unmarshal([]byte(memo), &packetMetadata)
	require.NoError(t, err)

	nextBz, err := json.Marshal(packetMetadata.Forward.Next)
	require.NoError(t, err)
	require.Equal(t, `{"forward":{"receiver":"noble1l505zhahp24v5jsmps9vs5asah759fdce06sfp","port":"transfer","channel":"channel-0","timeout":0}}`, string(nextBz))
}

func TestTimeoutUnmarshalString(t *testing.T) {
	const memo = "{\"forward\":{\"receiver\":\"noble1f4cur2krsua2th9kkp7n0zje4stea4p9tu70u8\",\"port\":\"transfer\",\"channel\":\"channel-0\",\"timeout\":\"60s\"}}"
	var packetMetadata types.PacketMetadata
	//memo := "trest"
	err := json.Unmarshal([]byte(memo), &packetMetadata)
	require.NoError(t, err)

	timeoutBz, err := json.Marshal(packetMetadata.Forward.Timeout)
	require.NoError(t, err)

	require.Equal(t, "60000000000", string(timeoutBz))
}

func TestTimeoutUnmarshalString2(t *testing.T) {
	const memo = "{\"wasm\":{\"contract\":\"neutron1mrm80xxdv8yhrt6gqvx2n638vjh23j023xj5yufha9y02gvskmaq6prr8z\",\"msg\":{\"swap_and_action\":{\"fee_swap\":{\"swap_venue_name\":\"neutron-astroport\",\"coin_out\":{\"denom\":\"untrn\",\"amount\":\"200000\"},\"operations\":[{\"pool\":\"neutron1u4v7xcvkhz8sxs3u9mjhprwc8vwc2p08x0tje4ugtrrkjhkagdysztt5dq\",\"denom_in\":\"ibc/376222D6D9DAE23092E29740E56B758580935A6D77C24C2ABD57A6A78A1F3955\",\"denom_out\":\"untrn\"}]},\"user_swap\":{\"swap_venue_name\":\"neutron-astroport\",\"operations\":[{\"pool\":\"neutron1u4v7xcvkhz8sxs3u9mjhprwc8vwc2p08x0tje4ugtrrkjhkagdysztt5dq\",\"denom_in\":\"ibc/376222D6D9DAE23092E29740E56B758580935A6D77C24C2ABD57A6A78A1F3955\",\"denom_out\":\"untrn\"},{\"pool\":\"neutron1e22zh5p8meddxjclevuhjmfj69jxfsa8uu3jvht72rv9d8lkhves6t8veq\",\"denom_in\":\"untrn\",\"denom_out\":\"ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9\"}]},\"min_coin\":{\"denom\":\"ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9\",\"amount\":\"361239\"},\"timeout_timestamp\":1690513310894078500,\"post_swap_action\":{\"ibc_transfer\":{\"ibc_info\":{\"source_channel\":\"channel-1\",\"receiver\":\"cosmos1aygdt8742gamxv8ca99wzh56ry4xw5s39smmhm\",\"fee\":{\"recv_fee\":[],\"ack_fee\":[{\"denom\":\"untrn\",\"amount\":\"100000\"}],\"timeout_fee\":[{\"denom\":\"untrn\",\"amount\":\"100000\"}]},\"memo\":\"\",\"recover_address\":\"neutron1aygdt8742gamxv8ca99wzh56ry4xw5s3p0jedu\"}}},\"affiliates\":[]}}}}"

	customJSON := &types.JSONObject{}
	//memo := "trest"
	err := json.Unmarshal([]byte(memo), customJSON)
	require.NoError(t, err)

	fmt.Println(customJSON)
}

func TestTimeoutUnmarshalJSON(t *testing.T) {
	const memo = "{\"forward\":{\"receiver\":\"noble1f4cur2krsua2th9kkp7n0zje4stea4p9tu70u8\",\"port\":\"transfer\",\"channel\":\"channel-0\",\"timeout\": 60000000000}}"
	var packetMetadata types.PacketMetadata

	err := json.Unmarshal([]byte(memo), &packetMetadata)
	require.NoError(t, err)

	timeoutBz, err := json.Marshal(packetMetadata.Forward.Timeout)
	require.NoError(t, err)

	require.Equal(t, "60000000000", string(timeoutBz))
}
