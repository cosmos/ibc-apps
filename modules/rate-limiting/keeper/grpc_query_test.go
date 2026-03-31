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

func (s *KeeperTestSuite) TestPaginatedQueries() {
	paginateRateLimits := func(fetch func(*querytypes.PageRequest) ([]types.RateLimit, *querytypes.PageResponse, error)) []types.RateLimit {
		firstPageItems, firstPage, err := fetch(&querytypes.PageRequest{Limit: 1})
		s.Require().NoError(err)
		s.Require().Len(firstPageItems, 1)
		s.Require().NotEmpty(firstPage.NextKey)

		secondPageItems, _, err := fetch(&querytypes.PageRequest{Limit: 10, Key: firstPage.NextKey})
		s.Require().NoError(err)
		s.Require().NotEmpty(secondPageItems)

		combined := append([]types.RateLimit{}, firstPageItems...)
		combined = append(combined, secondPageItems...)
		return combined
	}

	s.Run("all_rate_limits", func() {
		expectedRateLimits := s.setupQueryRateLimitTests()

		combined := paginateRateLimits(func(p *querytypes.PageRequest) ([]types.RateLimit, *querytypes.PageResponse, error) {
			resp, err := s.QueryClient.AllRateLimits(context.Background(), &types.QueryAllRateLimitsRequest{Pagination: p})
			if err != nil {
				return nil, nil, err
			}
			return resp.RateLimits, resp.Pagination, nil
		})
		s.Require().Len(combined, len(expectedRateLimits))
		s.Require().ElementsMatch(expectedRateLimits, combined)
	})

	s.Run("all_blacklisted_denoms", func() {
		s.App.RatelimitKeeper.AddDenomToBlacklist(s.Ctx, "denom-A")
		s.App.RatelimitKeeper.AddDenomToBlacklist(s.Ctx, "denom-B")

		firstPage, err := s.QueryClient.AllBlacklistedDenoms(context.Background(), &types.QueryAllBlacklistedDenomsRequest{
			Pagination: &querytypes.PageRequest{Limit: 1},
		})
		s.Require().NoError(err)
		s.Require().Len(firstPage.Denoms, 1)
		s.Require().NotEmpty(firstPage.Pagination.NextKey)

		secondPage, err := s.QueryClient.AllBlacklistedDenoms(context.Background(), &types.QueryAllBlacklistedDenomsRequest{
			Pagination: &querytypes.PageRequest{Limit: 10, Key: firstPage.Pagination.NextKey},
		})
		s.Require().NoError(err)
		s.Require().NotEmpty(secondPage.Denoms)

		combined := append([]string{}, firstPage.Denoms...)
		combined = append(combined, secondPage.Denoms...)
		s.Require().ElementsMatch([]string{"denom-A", "denom-B"}, combined)
	})

	s.Run("all_whitelisted_addresses", func() {
		expectedWhitelist := []types.WhitelistedAddressPair{
			{Sender: "address-A", Receiver: "address-B"},
			{Sender: "address-C", Receiver: "address-D"},
		}
		for _, pair := range expectedWhitelist {
			s.App.RatelimitKeeper.SetWhitelistedAddressPair(s.Ctx, pair)
		}

		firstPage, err := s.QueryClient.AllWhitelistedAddresses(context.Background(), &types.QueryAllWhitelistedAddressesRequest{
			Pagination: &querytypes.PageRequest{Limit: 1},
		})
		s.Require().NoError(err)
		s.Require().Len(firstPage.AddressPairs, 1)
		s.Require().NotEmpty(firstPage.Pagination.NextKey)

		secondPage, err := s.QueryClient.AllWhitelistedAddresses(context.Background(), &types.QueryAllWhitelistedAddressesRequest{
			Pagination: &querytypes.PageRequest{Limit: 10, Key: firstPage.Pagination.NextKey},
		})
		s.Require().NoError(err)
		s.Require().NotEmpty(secondPage.AddressPairs)

		combined := append([]types.WhitelistedAddressPair{}, firstPage.AddressPairs...)
		combined = append(combined, secondPage.AddressPairs...)
		s.Require().ElementsMatch(expectedWhitelist, combined)
	})

	addRateLimitWithChain := func(chainID, connectionID, channelID, denom string) {
		clientID := fmt.Sprintf("07-tendermint-%s", channelID)
		clientState := ibctmtypes.NewClientState(
			chainID, ibctmtypes.Fraction{}, time.Duration(0), time.Duration(0), time.Duration(0), clienttypes.Height{}, nil, nil,
		)
		connection := connectiontypes.ConnectionEnd{ClientId: clientID}
		channel := channeltypes.Channel{ConnectionHops: []string{connectionID}}

		s.App.IBCKeeper.ClientKeeper.SetClientState(s.Ctx, clientID, clientState)
		s.App.IBCKeeper.ConnectionKeeper.SetConnection(s.Ctx, connectionID, connection)
		s.App.IBCKeeper.ChannelKeeper.SetChannel(s.Ctx, transfertypes.PortID, channelID, channel)

		s.App.RatelimitKeeper.SetRateLimit(s.Ctx, types.RateLimit{
			Path: &types.Path{Denom: denom, ChannelOrClientId: channelID},
		})
	}

	s.Run("rate_limits_by_chain_id", func() {
		targetChainID := "chain-target"
		addRateLimitWithChain(targetChainID, "connection-chain-pg-1", "channel-chain-pg-1", "denom-a")
		addRateLimitWithChain(targetChainID, "connection-chain-pg-2", "channel-chain-pg-2", "denom-b")
		addRateLimitWithChain("chain-other", "connection-chain-pg-3", "channel-chain-pg-3", "denom-c")

		combined := paginateRateLimits(func(p *querytypes.PageRequest) ([]types.RateLimit, *querytypes.PageResponse, error) {
			resp, err := s.QueryClient.RateLimitsByChainId(context.Background(), &types.QueryRateLimitsByChainIdRequest{
				ChainId:    targetChainID,
				Pagination: p,
			})
			if err != nil {
				return nil, nil, err
			}
			return resp.RateLimits, resp.Pagination, nil
		})

		denoms := []string{combined[0].Path.Denom, combined[1].Path.Denom}
		s.Require().ElementsMatch([]string{"denom-a", "denom-b"}, denoms)
	})

	s.Run("rate_limits_by_channel_or_client_id", func() {
		targetChannelID := "channel-target"
		addRateLimitWithChain("chain-1", "connection-10", targetChannelID, "denom-a")
		// Same channel, different denom -> distinct rate limit entries.
		s.App.RatelimitKeeper.SetRateLimit(s.Ctx, types.RateLimit{
			Path: &types.Path{Denom: "denom-b", ChannelOrClientId: targetChannelID},
		})
		addRateLimitWithChain("chain-2", "connection-11", "channel-other", "denom-c")

		combined := paginateRateLimits(func(p *querytypes.PageRequest) ([]types.RateLimit, *querytypes.PageResponse, error) {
			resp, err := s.QueryClient.RateLimitsByChannelOrClientId(context.Background(), &types.QueryRateLimitsByChannelOrClientIdRequest{
				ChannelOrClientId: targetChannelID,
				Pagination:        p,
			})
			if err != nil {
				return nil, nil, err
			}
			return resp.RateLimits, resp.Pagination, nil
		})

		denoms := []string{combined[0].Path.Denom, combined[1].Path.Denom}
		s.Require().ElementsMatch([]string{"denom-a", "denom-b"}, denoms)
	})
}
