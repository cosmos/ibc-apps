package keeper_test

import "fmt"

func (s *KeeperTestSuite) TestPendingSendPacketPrefix() {
	// Store 5 packets across 4 channels
	channels := []string{"07-tendermint-1000", "07-tendermint-1005", "channel-1", "channel-11"}
	sendPackets := []string{}
	for _, channelId := range channels {
		for sequence := uint64(0); sequence < 5; sequence++ {
			err := s.App.RatelimitKeeper.SetPendingSendPacket(s.Ctx, channelId, sequence)
			s.Require().NoError(err, "unexpected error checking packet sent during current quota - channel %s, sequence %s", channelId, sequence)
			sendPackets = append(sendPackets, fmt.Sprintf("%s/%d", channelId, sequence))
		}
	}

	// Check that they each sequence number is found
	for _, channelId := range channels {
		for sequence := uint64(0); sequence < 5; sequence++ {
			found, err := s.App.RatelimitKeeper.CheckPacketSentDuringCurrentQuota(s.Ctx, channelId, sequence)
			s.Require().NoError(err, "unexpected error checking packet sent during current quota - channel %s, sequence %s", channelId, sequence)
			s.Require().True(found, "send packet should have been found - channel %s, sequence: %d", channelId, sequence)
		}
	}

	// Check lookup of all sequence numbers
	actualSendPackets, err := s.App.RatelimitKeeper.GetAllPendingSendPackets(s.Ctx)
	s.Require().NoError(err, "unexpected error getting pending send packets")
	s.Require().Equal(sendPackets, actualSendPackets, "all send packets")

	// Remove 0 sequence numbers and all sequence numbers from channel-0 + 07-tendermint-1005
	for _, channelId := range channels {
		err := s.App.RatelimitKeeper.RemovePendingSendPacket(s.Ctx, channelId, 0)
		s.Require().NoError(err, "unexpected error removing sequence 0 pending send packet - channel %s", channelId)
	}
	err = s.App.RatelimitKeeper.RemoveAllChannelPendingSendPackets(s.Ctx, "channel-1")
	s.Require().NoError(err, "unexpected error removing all pending send packets - channel %s", "channel-1")
	err = s.App.RatelimitKeeper.RemoveAllChannelPendingSendPackets(s.Ctx, "07-tendermint-1005")
	s.Require().NoError(err, "unexpected error removing all pending send packets - channel %s", "07-tendermint-1005")

	// Check that only the remaining sequences are found
	for _, channelId := range channels {
		for sequence := uint64(0); sequence < 5; sequence++ {
			removed := (channelId == "channel-1") || (channelId == "07-tendermint-1005") || (sequence == 0)
			actual, err := s.App.RatelimitKeeper.CheckPacketSentDuringCurrentQuota(s.Ctx, channelId, sequence)
			s.Require().NoError(err, "unexpected error checking packet sent during current quota - channel %s, sequence %s", channelId, sequence)

			// Assert that if we did not remove the packet, then we
			// successfully find it when checking the quota
			s.Require().Equal(!removed, actual, "send packet after removal - channel: %s, sequence: %d", channelId, sequence)
		}
	}
}
