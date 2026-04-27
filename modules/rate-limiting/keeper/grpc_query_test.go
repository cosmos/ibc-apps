package keeper_test

import (
	"context"
	"fmt"
	"time"

	querytypes "github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/ibc-apps/modules/rate-limiting/v10/types"

	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v10/modules/core/02-client/types"
	connectiontypes "github.com/cosmos/ibc-go/v10/modules/core/03-connection/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	ibctmtypes "github.com/cosmos/ibc-go/v10/modules/light-clients/07-tendermint"
)

// Add three rate limits on different channels
// Each should have a different chainId
func (s *KeeperTestSuite) setupQueryRateLimitTests() []types.RateLimit {
	rateLimits := []types.RateLimit{}
	for i := int64(0); i <= 2; i++ {
		clientId := fmt.Sprintf("07-tendermint-%d", i)
		chainId := fmt.Sprintf("chain-%d", i)
		connectionId := fmt.Sprintf("connection-%d", i)
		channelId := fmt.Sprintf("channel-%d", i)

		// First register the client, connection, and channel (so we can map back to chainId)
		// Nothing in the client state matters besides the chainId
		clientState := ibctmtypes.NewClientState(
			chainId, ibctmtypes.Fraction{}, time.Duration(0), time.Duration(0), time.Duration(0), clienttypes.Height{}, nil, nil,
		)
		connection := connectiontypes.ConnectionEnd{ClientId: clientId}
		channel := channeltypes.Channel{ConnectionHops: []string{connectionId}}

		s.App.IBCKeeper.ClientKeeper.SetClientState(s.Ctx, clientId, clientState)
		s.App.IBCKeeper.ConnectionKeeper.SetConnection(s.Ctx, connectionId, connection)
		s.App.IBCKeeper.ChannelKeeper.SetChannel(s.Ctx, transfertypes.PortID, channelId, channel)

		// Then add the rate limit
		rateLimit := types.RateLimit{
			Path: &types.Path{Denom: "denom", ChannelOrClientId: channelId},
		}
		s.App.RatelimitKeeper.SetRateLimit(s.Ctx, rateLimit)
		rateLimits = append(rateLimits, rateLimit)
	}
	return rateLimits
}

func (s *KeeperTestSuite) TestQueryAllRateLimits() {
	expectedRateLimits := s.setupQueryRateLimitTests()
	queryResponse, err := s.QueryClient.AllRateLimits(context.Background(), &types.QueryAllRateLimitsRequest{})
	s.Require().NoError(err)
	s.Require().ElementsMatch(expectedRateLimits, queryResponse.RateLimits)
}

func (s *KeeperTestSuite) TestQueryRateLimit() {
	allRateLimits := s.setupQueryRateLimitTests()
	for _, expectedRateLimit := range allRateLimits {
		queryResponse, err := s.QueryClient.RateLimit(context.Background(), &types.QueryRateLimitRequest{
			Denom:             expectedRateLimit.Path.Denom,
			ChannelOrClientId: expectedRateLimit.Path.ChannelOrClientId,
		})
		s.Require().NoError(err, "no error expected when querying rate limit on channel: %s", expectedRateLimit.Path.ChannelOrClientId)
		s.Require().Equal(expectedRateLimit, *queryResponse.RateLimit)
	}
}

func (s *KeeperTestSuite) TestQueryRateLimitsByChainId() {
	allRateLimits := s.setupQueryRateLimitTests()
	for i, expectedRateLimit := range allRateLimits {
		chainId := fmt.Sprintf("chain-%d", i)
		queryResponse, err := s.QueryClient.RateLimitsByChainId(context.Background(), &types.QueryRateLimitsByChainIdRequest{
			ChainId: chainId,
		})
		s.Require().NoError(err, "no error expected when querying rate limit on chain: %s", chainId)
		s.Require().Len(queryResponse.RateLimits, 1)
		s.Require().Equal(expectedRateLimit, queryResponse.RateLimits[0])
	}
}

func (s *KeeperTestSuite) TestQueryRateLimitsByChannelOrClientId() {
	allRateLimits := s.setupQueryRateLimitTests()
	for i, expectedRateLimit := range allRateLimits {
		channelId := fmt.Sprintf("channel-%d", i)
		queryResponse, err := s.QueryClient.RateLimitsByChannelOrClientId(context.Background(), &types.QueryRateLimitsByChannelOrClientIdRequest{
			ChannelOrClientId: channelId,
		})
		s.Require().NoError(err, "no error expected when querying rate limit on channel: %s", channelId)
		s.Require().Len(queryResponse.RateLimits, 1)
		s.Require().Equal(expectedRateLimit, queryResponse.RateLimits[0])
	}
}

func (s *KeeperTestSuite) TestQueryAllBlacklistedDenoms() {
	s.App.RatelimitKeeper.AddDenomToBlacklist(s.Ctx, "denom-A")
	s.App.RatelimitKeeper.AddDenomToBlacklist(s.Ctx, "denom-B")

	queryResponse, err := s.QueryClient.AllBlacklistedDenoms(context.Background(), &types.QueryAllBlacklistedDenomsRequest{})
	s.Require().NoError(err, "no error expected when querying blacklisted denoms")
	s.Require().Equal([]string{"denom-A", "denom-B"}, queryResponse.Denoms)
}

func (s *KeeperTestSuite) TestQueryAllWhitelistedAddresses() {
	s.App.RatelimitKeeper.SetWhitelistedAddressPair(s.Ctx, types.WhitelistedAddressPair{
		Sender:   "address-A",
		Receiver: "address-B",
	})
	s.App.RatelimitKeeper.SetWhitelistedAddressPair(s.Ctx, types.WhitelistedAddressPair{
		Sender:   "address-C",
		Receiver: "address-D",
	})
	queryResponse, err := s.QueryClient.AllWhitelistedAddresses(context.Background(), &types.QueryAllWhitelistedAddressesRequest{})
	s.Require().NoError(err, "no error expected when querying whitelisted addresses")

	expectedWhitelist := []types.WhitelistedAddressPair{
		{Sender: "address-A", Receiver: "address-B"},
		{Sender: "address-C", Receiver: "address-D"},
	}
	s.Require().Equal(expectedWhitelist, queryResponse.AddressPairs)
}

func (s *KeeperTestSuite) addChainRateLimit(chainID, channelID, denom string) {
	clientID := "07-tendermint-" + channelID
	connectionID := "connection-" + channelID
	s.App.IBCKeeper.ClientKeeper.SetClientState(s.Ctx, clientID, ibctmtypes.NewClientState(
		chainID, ibctmtypes.Fraction{}, 0, 0, 0, clienttypes.Height{}, nil, nil,
	))
	s.App.IBCKeeper.ConnectionKeeper.SetConnection(s.Ctx, connectionID, connectiontypes.ConnectionEnd{ClientId: clientID})
	s.App.IBCKeeper.ChannelKeeper.SetChannel(s.Ctx, transfertypes.PortID, channelID, channeltypes.Channel{ConnectionHops: []string{connectionID}})
	s.App.RatelimitKeeper.SetRateLimit(s.Ctx, types.RateLimit{
		Path: &types.Path{Denom: denom, ChannelOrClientId: channelID},
	})
}

func (s *KeeperTestSuite) TestPaginatedQueries() {
	s.Run("all_rate_limits", func() {
		s.SetupTest()
		expected := s.setupQueryRateLimitTests()

		first, err := s.QueryClient.AllRateLimits(context.Background(), &types.QueryAllRateLimitsRequest{
			Pagination: &querytypes.PageRequest{Limit: 1},
		})
		s.Require().NoError(err)
		s.Require().Len(first.RateLimits, 1)
		s.Require().NotEmpty(first.Pagination.NextKey)

		rest, err := s.QueryClient.AllRateLimits(context.Background(), &types.QueryAllRateLimitsRequest{
			Pagination: &querytypes.PageRequest{Key: first.Pagination.NextKey, Limit: 100},
		})
		s.Require().NoError(err)
		s.Require().ElementsMatch(expected, append(first.RateLimits, rest.RateLimits...))
	})

	s.Run("all_blacklisted_denoms", func() {
		s.SetupTest()
		expected := []string{"denom-A", "denom-B"}
		for _, d := range expected {
			s.App.RatelimitKeeper.AddDenomToBlacklist(s.Ctx, d)
		}

		first, err := s.QueryClient.AllBlacklistedDenoms(context.Background(), &types.QueryAllBlacklistedDenomsRequest{
			Pagination: &querytypes.PageRequest{Limit: 1},
		})
		s.Require().NoError(err)
		s.Require().Len(first.Denoms, 1)
		s.Require().NotEmpty(first.Pagination.NextKey)

		rest, err := s.QueryClient.AllBlacklistedDenoms(context.Background(), &types.QueryAllBlacklistedDenomsRequest{
			Pagination: &querytypes.PageRequest{Key: first.Pagination.NextKey, Limit: 100},
		})
		s.Require().NoError(err)
		s.Require().ElementsMatch(expected, append(first.Denoms, rest.Denoms...))
	})

	s.Run("all_whitelisted_addresses", func() {
		s.SetupTest()
		expected := []types.WhitelistedAddressPair{
			{Sender: "address-A", Receiver: "address-B"},
			{Sender: "address-C", Receiver: "address-D"},
		}
		for _, pair := range expected {
			s.App.RatelimitKeeper.SetWhitelistedAddressPair(s.Ctx, pair)
		}

		first, err := s.QueryClient.AllWhitelistedAddresses(context.Background(), &types.QueryAllWhitelistedAddressesRequest{
			Pagination: &querytypes.PageRequest{Limit: 1},
		})
		s.Require().NoError(err)
		s.Require().Len(first.AddressPairs, 1)
		s.Require().NotEmpty(first.Pagination.NextKey)

		rest, err := s.QueryClient.AllWhitelistedAddresses(context.Background(), &types.QueryAllWhitelistedAddressesRequest{
			Pagination: &querytypes.PageRequest{Key: first.Pagination.NextKey, Limit: 100},
		})
		s.Require().NoError(err)
		s.Require().ElementsMatch(expected, append(first.AddressPairs, rest.AddressPairs...))
	})

	s.Run("rate_limits_by_chain_id", func() {
		s.SetupTest()
		const target = "chain-target"
		s.addChainRateLimit(target, "channel-1", "denom-a")
		s.addChainRateLimit(target, "channel-2", "denom-b")
		s.addChainRateLimit("chain-other", "channel-3", "denom-c")

		first, err := s.QueryClient.RateLimitsByChainId(context.Background(), &types.QueryRateLimitsByChainIdRequest{
			ChainId:    target,
			Pagination: &querytypes.PageRequest{Limit: 1},
		})
		s.Require().NoError(err)
		s.Require().Len(first.RateLimits, 1)
		s.Require().NotEmpty(first.Pagination.NextKey)

		rest, err := s.QueryClient.RateLimitsByChainId(context.Background(), &types.QueryRateLimitsByChainIdRequest{
			ChainId:    target,
			Pagination: &querytypes.PageRequest{Key: first.Pagination.NextKey, Limit: 100},
		})
		s.Require().NoError(err)

		got := append(first.RateLimits, rest.RateLimits...)
		denoms := []string{got[0].Path.Denom, got[1].Path.Denom}
		s.Require().ElementsMatch([]string{"denom-a", "denom-b"}, denoms)
	})

	s.Run("rate_limits_by_channel_or_client_id", func() {
		s.SetupTest()
		const target = "channel-target"
		s.addChainRateLimit("chain-1", target, "denom-a")
		// Same channel, different denom -> distinct rate-limit entry.
		s.App.RatelimitKeeper.SetRateLimit(s.Ctx, types.RateLimit{
			Path: &types.Path{Denom: "denom-b", ChannelOrClientId: target},
		})
		s.addChainRateLimit("chain-2", "channel-other", "denom-c")

		first, err := s.QueryClient.RateLimitsByChannelOrClientId(context.Background(), &types.QueryRateLimitsByChannelOrClientIdRequest{
			ChannelOrClientId: target,
			Pagination:        &querytypes.PageRequest{Limit: 1},
		})
		s.Require().NoError(err)
		s.Require().Len(first.RateLimits, 1)
		s.Require().NotEmpty(first.Pagination.NextKey)

		rest, err := s.QueryClient.RateLimitsByChannelOrClientId(context.Background(), &types.QueryRateLimitsByChannelOrClientIdRequest{
			ChannelOrClientId: target,
			Pagination:        &querytypes.PageRequest{Key: first.Pagination.NextKey, Limit: 100},
		})
		s.Require().NoError(err)

		got := append(first.RateLimits, rest.RateLimits...)
		denoms := []string{got[0].Path.Denom, got[1].Path.Denom}
		s.Require().ElementsMatch([]string{"denom-a", "denom-b"}, denoms)
	})

	s.Run("count_total_omitted_for_key_resumed_pages", func() {
		s.SetupTest()
		const target = "channel-total-target"
		s.App.RatelimitKeeper.SetRateLimit(s.Ctx, types.RateLimit{Path: &types.Path{Denom: "denom-a", ChannelOrClientId: target}})
		s.App.RatelimitKeeper.SetRateLimit(s.Ctx, types.RateLimit{Path: &types.Path{Denom: "denom-b", ChannelOrClientId: target}})
		s.App.RatelimitKeeper.SetRateLimit(s.Ctx, types.RateLimit{Path: &types.Path{Denom: "denom-c", ChannelOrClientId: "channel-other"}})

		first, err := s.QueryClient.RateLimitsByChannelOrClientId(context.Background(), &types.QueryRateLimitsByChannelOrClientIdRequest{
			ChannelOrClientId: target,
			Pagination:        &querytypes.PageRequest{Limit: 1, CountTotal: true},
		})
		s.Require().NoError(err)
		s.Require().Equal(uint64(2), first.Pagination.Total, "full-scan page should report Total")

		second, err := s.QueryClient.RateLimitsByChannelOrClientId(context.Background(), &types.QueryRateLimitsByChannelOrClientIdRequest{
			ChannelOrClientId: target,
			Pagination:        &querytypes.PageRequest{Key: first.Pagination.NextKey, Limit: 1, CountTotal: true},
		})
		s.Require().NoError(err)
		s.Require().Equal(uint64(0), second.Pagination.Total, "key-resumed page should omit Total")
	})

	s.Run("count_total_omitted_when_scan_budget_hits", func() {
		s.SetupTest()
		for i := 0; i <= 5000; i++ {
			s.App.RatelimitKeeper.SetRateLimit(s.Ctx, types.RateLimit{
				Path: &types.Path{Denom: fmt.Sprintf("denom-scan-%d", i), ChannelOrClientId: fmt.Sprintf("channel-scan-%d", i)},
			})
		}

		resp, err := s.QueryClient.RateLimitsByChannelOrClientId(context.Background(), &types.QueryRateLimitsByChannelOrClientIdRequest{
			ChannelOrClientId: "channel-missing",
			Pagination:        &querytypes.PageRequest{Limit: 1, CountTotal: true},
		})
		s.Require().NoError(err)
		s.Require().Empty(resp.RateLimits)
		s.Require().NotEmpty(resp.Pagination.NextKey, "scan-cap hit should surface NextKey")
		s.Require().Equal(uint64(0), resp.Pagination.Total, "scan-cap-truncated page should omit Total")
	})
}
