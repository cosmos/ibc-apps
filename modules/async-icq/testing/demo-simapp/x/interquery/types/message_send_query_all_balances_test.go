package types

import (
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/ibc-apps/modules/async-icq/v7/interchain-query-demo/testutil/sample"
	"github.com/stretchr/testify/require"
)

func TestMsgSendQueryAllBalances_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgSendQueryAllBalances
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgSendQueryAllBalances{
				Creator: "invalid_address",
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "valid address",
			msg: MsgSendQueryAllBalances{
				Creator: sample.AccAddress(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
				return
			}
			require.NoError(t, err)
		})
	}
}
