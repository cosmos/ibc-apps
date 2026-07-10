package keeper_test

import (
	"fmt"
	"strings"

	"github.com/cosmos/ibc-apps/modules/rate-limiting/v10/types"
)

const (
	pendingPacketChannelToRemove = "channel-1"
	pendingPacketClientToRemove  = "07-tendermint-1005"
	pendingPacketDenomA          = "denom-a"
	pendingPacketDenomB          = "denom-b"
	pendingPacketDenomErr        = "pending packet denom must be specified"
)

func (s *KeeperTestSuite) TestPendingSendPacketPrefix() {
	// Store 5 packets across 4 channels and 2 denoms
	channels := []string{"07-tendermint-1000", pendingPacketClientToRemove, pendingPacketChannelToRemove, "channel-11"}
	denoms := []string{pendingPacketDenomA, pendingPacketDenomB}
	sendPackets := []string{}
	for _, channelId := range channels {
		for _, denom := range denoms {
			for sequence := uint64(0); sequence < 5; sequence++ {
				err := s.App.RatelimitKeeper.SetPendingSendPacket(s.Ctx, channelId, sequence, denom)
				s.Require().NoError(err, "unexpected error setting pending send packet sequence - channel %s, sequence %d, denom %s", channelId, sequence, denom)
				sendPackets = append(sendPackets, fmt.Sprintf("%s/%d/%s", channelId, sequence, denom))
			}
		}
	}

	// Check that each sequence number is found
	for _, channelId := range channels {
		for _, denom := range denoms {
			for sequence := uint64(0); sequence < 5; sequence++ {
				found, err := s.App.RatelimitKeeper.CheckPacketSentDuringCurrentQuota(s.Ctx, channelId, sequence, denom)
				s.Require().NoError(err, "unexpected error checking packet sent during current quota - channel %s, sequence %d, denom %s", channelId, sequence, denom)
				s.Require().True(found, "send packet should have been found - channel %s, sequence: %d, denom: %s", channelId, sequence, denom)
			}
		}
	}

	// Check lookup of all sequence numbers
	actualSendPackets, err := s.App.RatelimitKeeper.GetAllPendingSendPackets(s.Ctx)
	s.Require().NoError(err, "unexpected error getting pending send packets")
	s.Require().ElementsMatch(sendPackets, actualSendPackets, "all send packets")

	// Remove denom-a sequence 0 and all denom-scoped sequence numbers from channel-1 + 07-tendermint-1005
	for _, channelId := range channels {
		err = s.App.RatelimitKeeper.RemovePendingSendPacket(s.Ctx, channelId, 0, pendingPacketDenomA)
		s.Require().NoError(err, "unexpected error removing pending send packet - channel %s, sequence 0", channelId)
	}
	err = s.App.RatelimitKeeper.RemoveAllChannelPendingSendPackets(s.Ctx, pendingPacketChannelToRemove, pendingPacketDenomA)
	s.Require().NoError(err, "unexpected error removing all pending send packets - channel %s", pendingPacketChannelToRemove)
	err = s.App.RatelimitKeeper.RemoveAllChannelPendingSendPackets(s.Ctx, pendingPacketClientToRemove, pendingPacketDenomB)
	s.Require().NoError(err, "unexpected error removing all pending send packets - channel %s", pendingPacketClientToRemove)

	// Check that only the remaining sequences are found
	for _, channelId := range channels {
		for _, denom := range denoms {
			for sequence := uint64(0); sequence < 5; sequence++ {
				removed := (denom == pendingPacketDenomA && sequence == 0) || (channelId == pendingPacketChannelToRemove && denom == pendingPacketDenomA) || (channelId == pendingPacketClientToRemove && denom == pendingPacketDenomB)
				actual, err := s.App.RatelimitKeeper.CheckPacketSentDuringCurrentQuota(s.Ctx, channelId, sequence, denom)
				s.Require().NoError(err, "unexpected error checking packet sent during current quota - channel %s, sequence %d, denom %s", channelId, sequence, denom)

				// Assert that if we did not remove the packet, then we
				// successfully find it when checking the quota
				s.Require().Equal(!removed, actual, "send packet after removal - channel: %s, sequence: %d, denom: %s", channelId, sequence, denom)
			}
		}
	}
}

func (s *KeeperTestSuite) TestPendingReceivePacketPrefix() {
	// Store 5 packets across 4 channels and 2 denoms
	channels := []string{"07-tendermint-1000", pendingPacketClientToRemove, pendingPacketChannelToRemove, "channel-11"}
	denoms := []string{pendingPacketDenomA, pendingPacketDenomB}
	receivePackets := []string{}
	for _, channelId := range channels {
		for _, denom := range denoms {
			for sequence := uint64(0); sequence < 5; sequence++ {
				err := s.App.RatelimitKeeper.SetPendingReceivePacket(s.Ctx, channelId, sequence, denom)
				s.Require().NoError(err, "unexpected error setting pending receive packet sequence - channel %s, sequence %d, denom %s", channelId, sequence, denom)
				receivePackets = append(receivePackets, fmt.Sprintf("%s/%d/%s", channelId, sequence, denom))
			}
		}
	}

	// Check that each sequence number is found
	for _, channelId := range channels {
		for _, denom := range denoms {
			for sequence := uint64(0); sequence < 5; sequence++ {
				found, err := s.App.RatelimitKeeper.CheckPacketReceivedDuringCurrentQuota(s.Ctx, channelId, sequence, denom)
				s.Require().NoError(err, "unexpected error checking packet received during current quota - channel %s, sequence %d, denom %s", channelId, sequence, denom)
				s.Require().True(found, "receive packet should have been found - channel %s, sequence: %d, denom: %s", channelId, sequence, denom)
			}
		}
	}

	// Check lookup of all sequence numbers
	actualReceivePackets, err := s.App.RatelimitKeeper.GetAllPendingReceivePackets(s.Ctx)
	s.Require().NoError(err, "unexpected error getting pending receive packets")
	s.Require().ElementsMatch(receivePackets, actualReceivePackets, "all receive packets")

	// Remove denom-a sequence 0 and all denom-scoped sequence numbers from channel-1 + 07-tendermint-1005
	for _, channelId := range channels {
		err := s.App.RatelimitKeeper.RemovePendingReceivePacket(s.Ctx, channelId, 0, pendingPacketDenomA)
		s.Require().NoError(err, "unexpected error removing pending receive packet - channel %s, sequence 0", channelId)
	}
	err = s.App.RatelimitKeeper.RemoveAllChannelPendingReceivePackets(s.Ctx, pendingPacketChannelToRemove, pendingPacketDenomA)
	s.Require().NoError(err, "unexpected error removing all pending receive packets - channel %s", pendingPacketChannelToRemove)
	err = s.App.RatelimitKeeper.RemoveAllChannelPendingReceivePackets(s.Ctx, pendingPacketClientToRemove, pendingPacketDenomB)
	s.Require().NoError(err, "unexpected error removing all pending receive packets - channel %s", pendingPacketClientToRemove)

	// Check that only the remaining sequences are found
	for _, channelId := range channels {
		for _, denom := range denoms {
			for sequence := uint64(0); sequence < 5; sequence++ {
				removed := (denom == pendingPacketDenomA && sequence == 0) || (channelId == pendingPacketChannelToRemove && denom == pendingPacketDenomA) || (channelId == pendingPacketClientToRemove && denom == pendingPacketDenomB)
				actual, err := s.App.RatelimitKeeper.CheckPacketReceivedDuringCurrentQuota(s.Ctx, channelId, sequence, denom)
				s.Require().NoError(err, "unexpected error checking packet received during current quota - channel %s, sequence %d, denom %s", channelId, sequence, denom)

				// Assert that if we did not remove the packet, then we
				// successfully find it when checking the quota
				s.Require().Equal(!removed, actual, "receive packet after removal - channel: %s, sequence: %d, denom: %s", channelId, sequence, denom)
			}
		}
	}
}

func (s *KeeperTestSuite) TestPendingPacketValidation() {
	longChannelId := strings.Repeat("a", types.PendingSendPacketChannelLength+1)

	testCases := []struct {
		name      string
		call      func() error
		expErrMsg string
	}{
		{
			name: "set send empty denom",
			call: func() error {
				return s.App.RatelimitKeeper.SetPendingSendPacket(s.Ctx, channelId, 1, "")
			},
			expErrMsg: pendingPacketDenomErr,
		},
		{
			name: "set receive invalid channel",
			call: func() error {
				return s.App.RatelimitKeeper.SetPendingReceivePacket(s.Ctx, longChannelId, 1, denom)
			},
			expErrMsg: "greater than the allowed length 64",
		},
		{
			name: "set send invalid channel delimiter",
			call: func() error {
				return s.App.RatelimitKeeper.SetPendingSendPacket(s.Ctx, "channel-\x00", 1, denom)
			},
			expErrMsg: "cannot contain 0x00",
		},
		{
			name: "remove send empty denom",
			call: func() error {
				return s.App.RatelimitKeeper.RemovePendingSendPacket(s.Ctx, channelId, 1, "")
			},
			expErrMsg: pendingPacketDenomErr,
		},
		{
			name: "remove receive invalid denom delimiter",
			call: func() error {
				return s.App.RatelimitKeeper.RemovePendingReceivePacket(s.Ctx, channelId, 1, "denom\x00")
			},
			expErrMsg: "pending packet denom cannot contain 0x00",
		},
		{
			name: "check receive empty denom",
			call: func() error {
				_, err := s.App.RatelimitKeeper.CheckPacketReceivedDuringCurrentQuota(s.Ctx, channelId, 1, "")
				return err
			},
			expErrMsg: pendingPacketDenomErr,
		},
		{
			name: "remove all send empty denom",
			call: func() error {
				return s.App.RatelimitKeeper.RemoveAllChannelPendingSendPackets(s.Ctx, channelId, "")
			},
			expErrMsg: pendingPacketDenomErr,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			err := tc.call()
			s.Require().ErrorContains(err, tc.expErrMsg)
		})
	}
}
