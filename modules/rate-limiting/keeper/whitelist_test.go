package keeper_test

import "github.com/cosmos/ibc-apps/modules/rate-limiting/v7/types"

func (s *KeeperTestSuite) TestAddressWhitelist() {
	// Store addresses in whitelist
	expectedWhitelist := []types.WhitelistedAddressPair{
		{Sender: "sender-1", Receiver: "receiver-1"},
		{Sender: "sender-2", Receiver: "receiver-2"},
		{Sender: "sender-3", Receiver: "receiver-3"},
	}
	for _, addressPair := range expectedWhitelist {
		s.App.RatelimitKeeper.SetWhitelistedAddressPair(s.Ctx, addressPair)
	}

	// Confirm that each was found
	for _, addressPair := range expectedWhitelist {
		found := s.App.RatelimitKeeper.IsAddressPairWhitelisted(s.Ctx, addressPair.Sender, addressPair.Receiver)
		s.Require().True(found, "address pair should have been whitelisted (%s/%s)",
			addressPair.Sender, addressPair.Receiver)
	}

	// Confirm that looking both the sender and receiver must match for the pair to be whitelisted
	for _, addressPair := range expectedWhitelist {
		found := s.App.RatelimitKeeper.IsAddressPairWhitelisted(s.Ctx, addressPair.Sender, "fake-receiver")
		s.Require().False(found, "address pair should not have been whitelisted (%s/%s)",
			addressPair.Sender, "fake-receiver")

		found = s.App.RatelimitKeeper.IsAddressPairWhitelisted(s.Ctx, "fake-sender", addressPair.Receiver)
		s.Require().False(found, "address pair should not have been whitelisted (%s/%s)",
			"fake-sender", addressPair.Receiver)
	}

	// Check GetAll
	actualWhitelist := s.App.RatelimitKeeper.GetAllWhitelistedAddressPairs(s.Ctx)
	s.Require().Equal(expectedWhitelist, actualWhitelist, "whitelist get all")

	// Finally, remove each from whitelist
	for _, addressPair := range expectedWhitelist {
		s.App.RatelimitKeeper.RemoveWhitelistedAddressPair(s.Ctx, addressPair.Sender, addressPair.Receiver)
	}

	// Confirm there are no longer any whitelisted pairs
	actualWhitelist = s.App.RatelimitKeeper.GetAllWhitelistedAddressPairs(s.Ctx)
	s.Require().Empty(actualWhitelist, "whitelist should have been cleared")

	for _, addressPair := range expectedWhitelist {
		found := s.App.RatelimitKeeper.IsAddressPairWhitelisted(s.Ctx, addressPair.Sender, addressPair.Receiver)
		s.Require().False(found, "address pair should no longer be whitelisted (%s/%s)",
			addressPair.Sender, addressPair.Receiver)
	}
}
