// trace.go lifted from https://github.com/cosmos/ibc-go/blob/a4ef5360b49ad2118e1d68f25f13935162660e0b/modules/apps/transfer/types/trace.go#L31 https://github.com/cosmos/ibc-go/blob/a4ef5360b49ad2118e1d68f25f13935162660e0b/modules/apps/transfer/types/trace.go#L31 https://github.com/cosmos/ibc-go/blob/a4ef5360b49ad2118e1d68f25f13935162660e0b/modules/apps/transfer/types/trace.go#L31 https://github.com/cosmos/ibc-go/blob/a4ef5360b49ad2118e1d68f25f13935162660e0b/modules/apps/transfer/types/trace.go#L31
// to maintain backwards compatibility. may not be needed if the protocol has changed, but I can't tell that from the code
// right now.

package keeper

import (
	"crypto/sha256"
	"fmt"
	cmtbytes "github.com/cometbft/cometbft/libs/bytes"
	"strings"

	channeltypes "github.com/cosmos/ibc-go/v9/modules/core/04-channel/types"
)

const denomPrefix = "ibc"

// DenomTrace contains the base denomination for ICS20 fungible tokens and the
// source tracing information path.
type denomTrace struct {
	// path defines the chain of port/channel identifiers used for tracing the
	// source of the fungible token.
	path string
	// base denomination of the relayed fungible token.
	baseDenom string
}

// ParseDenomTrace parses a string with the ibc prefix (denom trace) and the base denomination
// into a DenomTrace type.
//
// Examples:
//
// - "portidone/channel-0/uatom" => DenomTrace{Path: "portidone/channel-0", BaseDenom: "uatom"}
// - "portidone/channel-0/portidtwo/channel-1/uatom" => DenomTrace{Path: "portidone/channel-0/portidtwo/channel-1", BaseDenom: "uatom"}
// - "portidone/channel-0/gamm/pool/1" => DenomTrace{Path: "portidone/channel-0", BaseDenom: "gamm/pool/1"}
// - "gamm/pool/1" => DenomTrace{Path: "", BaseDenom: "gamm/pool/1"}
// - "uatom" => DenomTrace{Path: "", BaseDenom: "uatom"}
func parseDenomTrace(rawDenom string) denomTrace {
	denomSplit := strings.Split(rawDenom, "/")

	if denomSplit[0] == rawDenom {
		return denomTrace{
			path:      "",
			baseDenom: rawDenom,
		}
	}

	path, baseDenom := extractPathAndBaseFromFullDenom(denomSplit)
	return denomTrace{
		path:      path,
		baseDenom: baseDenom,
	}
}

// extractPathAndBaseFromFullDenom returns the trace path and the base denom from
// the elements that constitute the complete denom.
func extractPathAndBaseFromFullDenom(fullDenomItems []string) (string, string) {
	var (
		pathSlice      []string
		baseDenomSlice []string
	)

	length := len(fullDenomItems)
	for i := 0; i < length; i += 2 {
		// The IBC specification does not guarantee the expected format of the
		// destination port or destination channel identifier. A short term solution
		// to determine base denomination is to expect the channel identifier to be the
		// one ibc-go specifies. A longer term solution is to separate the path and base
		// denomination in the ICS20 packet. If an intermediate hop prefixes the full denom
		// with a channel identifier format different from our own, the base denomination
		// will be incorrectly parsed, but the token will continue to be treated correctly
		// as an IBC denomination. The hash used to store the token internally on our chain
		// will be the same value as the base denomination being correctly parsed.
		if i < length-1 && length > 2 && channeltypes.IsValidChannelID(fullDenomItems[i+1]) {
			pathSlice = append(pathSlice, fullDenomItems[i], fullDenomItems[i+1])
		} else {
			baseDenomSlice = fullDenomItems[i:]
			break
		}
	}

	path := strings.Join(pathSlice, "/")
	baseDenom := strings.Join(baseDenomSlice, "/")

	return path, baseDenom
}

// Hash returns the hex bytes of the SHA256 hash of the DenomTrace fields using the following formula:
//
// hash = sha256(tracePath + "/" + baseDenom)
func (dt denomTrace) hash() cmtbytes.HexBytes {
	hash := sha256.Sum256([]byte(dt.GetFullDenomPath()))
	return hash[:]
}

// GetPrefix returns the receiving denomination prefix composed by the trace info and a separator.
func (dt denomTrace) getPrefix() string {
	return dt.path + "/"
}

// IBCDenom a coin denomination for an ICS20 fungible token in the format
// 'ibc/{hash(tracePath + baseDenom)}'. If the trace is empty, it will return the base denomination.
func (dt denomTrace) IBCDenom() string {
	if dt.path != "" {
		return fmt.Sprintf("%s/%s", denomPrefix, dt.hash())
	}
	return dt.baseDenom
}

// GetFullDenomPath returns the full denomination according to the ICS20 specification:
// tracePath + "/" + baseDenom
// If there exists no trace then the base denomination is returned.
func (dt denomTrace) GetFullDenomPath() string {
	if dt.path == "" {
		return dt.baseDenom
	}
	return dt.getPrefix() + dt.baseDenom
}
