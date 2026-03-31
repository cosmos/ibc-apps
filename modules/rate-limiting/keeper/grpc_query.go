package keeper

import (
	"context"

	"github.com/cosmos/ibc-apps/modules/rate-limiting/v10/types"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/store/prefix"

	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	ibctmtypes "github.com/cosmos/ibc-go/v10/modules/light-clients/07-tendermint"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = Keeper{}

const (
	maxFilteredQueryScanCap = uint64(5000)
	maxPageLimit            = uint64(500)
	defaultPageLimit        = uint64(100)
)

// Query all rate limits
func (k Keeper) AllRateLimits(c context.Context, req *types.QueryAllRateLimitsRequest) (*types.QueryAllRateLimitsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	store := k.prefixedStore(ctx, types.RateLimitKeyPrefix)

	rateLimits := []types.RateLimit{}
	pageRes, err := query.Paginate(store, req.Pagination, func(_, value []byte) error {
		rateLimit := types.RateLimit{}
		k.cdc.MustUnmarshal(value, &rateLimit)
		rateLimits = append(rateLimits, rateLimit)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &types.QueryAllRateLimitsResponse{RateLimits: rateLimits, Pagination: pageRes}, nil
}

// Query a rate limit by denom and channelId
func (k Keeper) RateLimit(c context.Context, req *types.QueryRateLimitRequest) (*types.QueryRateLimitResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	rateLimit, found := k.GetRateLimit(ctx, req.Denom, req.ChannelOrClientId)
	if !found {
		return &types.QueryRateLimitResponse{}, nil
	}
	return &types.QueryRateLimitResponse{RateLimit: &rateLimit}, nil
}

// Query all rate limits for a given chain
func (k Keeper) RateLimitsByChainId(c context.Context, req *types.QueryRateLimitsByChainIdRequest) (*types.QueryRateLimitsByChainIdResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.ChainId == "" {
		return nil, status.Error(codes.InvalidArgument, "chain_id cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(c)
	store := k.prefixedStore(ctx, types.RateLimitKeyPrefix)

	rateLimits, pageRes, err := k.paginateFilteredRateLimits(store, req.Pagination, func(rateLimit types.RateLimit) (bool, error) {
		// Determine the client state from the channel Id
		_, clientState, err := k.channelKeeper.GetChannelClientState(ctx, transfertypes.PortID, rateLimit.Path.ChannelOrClientId)
		if err != nil {
			var ok bool
			clientState, ok = k.clientKeeper.GetClientState(ctx, rateLimit.Path.ChannelOrClientId)
			if !ok {
				return false, errorsmod.Wrapf(types.ErrInvalidClientState, "unable to fetch client state from channel or client id")
			}
		}
		client, ok := clientState.(*ibctmtypes.ClientState)
		if !ok {
			// If the client state is not a tendermint client state, we don't return the rate limit from this query
			return false, nil
		}
		return client.ChainId == req.ChainId, nil
	})
	if err != nil {
		return nil, err
	}

	return &types.QueryRateLimitsByChainIdResponse{RateLimits: rateLimits, Pagination: pageRes}, nil
}

// Query all rate limits for a given channel
func (k Keeper) RateLimitsByChannelOrClientId(c context.Context, req *types.QueryRateLimitsByChannelOrClientIdRequest) (*types.QueryRateLimitsByChannelOrClientIdResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.ChannelOrClientId == "" {
		return nil, status.Error(codes.InvalidArgument, "channel_or_client_id cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(c)
	store := k.prefixedStore(ctx, types.RateLimitKeyPrefix)

	rateLimits, pageRes, err := k.paginateFilteredRateLimits(store, req.Pagination, func(rateLimit types.RateLimit) (bool, error) {
		// If the channel ID matches, add the rate limit to the returned list
		return rateLimit.Path.ChannelOrClientId == req.ChannelOrClientId, nil
	})
	if err != nil {
		return nil, err
	}

	return &types.QueryRateLimitsByChannelOrClientIdResponse{RateLimits: rateLimits, Pagination: pageRes}, nil
}

// Query all blacklisted denoms
func (k Keeper) AllBlacklistedDenoms(c context.Context, req *types.QueryAllBlacklistedDenomsRequest) (*types.QueryAllBlacklistedDenomsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	store := k.prefixedStore(ctx, types.DenomBlacklistKeyPrefix)

	blacklistedDenoms := []string{}
	pageRes, err := query.Paginate(store, req.Pagination, func(key, _ []byte) error {
		blacklistedDenoms = append(blacklistedDenoms, string(key))
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &types.QueryAllBlacklistedDenomsResponse{Denoms: blacklistedDenoms, Pagination: pageRes}, nil
}

// Query all whitelisted addresses
func (k Keeper) AllWhitelistedAddresses(c context.Context, req *types.QueryAllWhitelistedAddressesRequest) (*types.QueryAllWhitelistedAddressesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	store := k.prefixedStore(ctx, types.AddressWhitelistKeyPrefix)

	whitelistedAddresses := []types.WhitelistedAddressPair{}
	pageRes, err := query.Paginate(store, req.Pagination, func(_, value []byte) error {
		whitelist := types.WhitelistedAddressPair{}
		k.cdc.MustUnmarshal(value, &whitelist)
		whitelistedAddresses = append(whitelistedAddresses, whitelist)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &types.QueryAllWhitelistedAddressesResponse{AddressPairs: whitelistedAddresses, Pagination: pageRes}, nil
}

func (k Keeper) paginateFilteredRateLimits(
	store prefix.Store,
	pagination *query.PageRequest,
	matchFn func(types.RateLimit) (bool, error),
) ([]types.RateLimit, *query.PageResponse, error) {
	if pagination != nil && (pagination.Offset > 0 || pagination.CountTotal || pagination.Reverse) {
		return nil, nil, status.Error(codes.InvalidArgument, "offset, count_total, and reverse are not supported for filtered queries")
	}

	limit := defaultPageLimit
	if pagination != nil && pagination.Limit != 0 {
		limit = pagination.Limit
	}
	if limit > maxPageLimit {
		limit = maxPageLimit
	}
	scanBudget := maxFilteredQueryScanCap
	var startKey []byte
	if pagination != nil {
		startKey = pagination.Key
	}

	iterator := store.Iterator(startKey, nil)
	defer iterator.Close()

	rateLimits := make([]types.RateLimit, 0, limit)
	var scanned uint64

	for ; iterator.Valid(); iterator.Next() {
		if scanned >= scanBudget {
			return rateLimits, &query.PageResponse{NextKey: append([]byte(nil), iterator.Key()...)}, nil
		}
		scanned++

		rateLimit := types.RateLimit{}
		k.cdc.MustUnmarshal(iterator.Value(), &rateLimit)

		matches, err := matchFn(rateLimit)
		if err != nil {
			return nil, nil, err
		}
		if !matches {
			continue
		}

		rateLimits = append(rateLimits, rateLimit)
		if uint64(len(rateLimits)) == limit {
			iterator.Next()
			if iterator.Valid() {
				return rateLimits, &query.PageResponse{NextKey: append([]byte(nil), iterator.Key()...)}, nil
			}
			break
		}
	}

	return rateLimits, &query.PageResponse{}, nil
}

func (k Keeper) prefixedStore(ctx sdk.Context, keyPrefix []byte) prefix.Store {
	adapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	return prefix.NewStore(adapter, keyPrefix)
}
