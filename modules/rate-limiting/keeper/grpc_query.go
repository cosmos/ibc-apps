package keeper

import (
	"context"

	"github.com/cosmos/ibc-apps/modules/rate-limiting/v10/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	ibctmtypes "github.com/cosmos/ibc-go/v10/modules/light-clients/07-tendermint"
)

var _ types.QueryServer = Keeper{}

const (
	maxFilteredQueryScanCap = uint64(5000)
	maxPageLimit            = uint64(500)
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

// RateLimitsByChainId returns rate limits whose channel resolves to the given chain.
// Filtered, paginated query: limit ≤ maxPageLimit, scan capped at maxFilteredQueryScanCap.
// NextKey is resumable; Total is populated only on a full prefix scan (no resume key
// and no scan-cap truncation).
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

// RateLimitsByChannelOrClientId returns rate limits for the given channel or client ID.
// Same pagination contract as RateLimitsByChainId.
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

// paginateFilteredRateLimits forks query.FilteredPaginate with two guards:
// rejects Limit > maxPageLimit, and stops after maxFilteredQueryScanCap
// iterations, surfacing the next key as PageResponse.NextKey for resumption.
func (k Keeper) paginateFilteredRateLimits(
	store prefix.Store,
	pagination *query.PageRequest,
	matchFn func(types.RateLimit) (bool, error),
) ([]types.RateLimit, *query.PageResponse, error) {
	if pagination != nil && pagination.Limit > maxPageLimit {
		return nil, nil, status.Errorf(codes.InvalidArgument, "limit must not exceed %d", maxPageLimit)
	}

	pageReq := initFilteredPageRequest(pagination)
	if pageReq.Offset > 0 && pageReq.Key != nil {
		return nil, nil, status.Error(codes.InvalidArgument, "either offset or key is expected, got both")
	}

	iterator := openFilteredIterator(store, pageReq.Key, pageReq.Reverse)
	defer iterator.Close()

	var (
		rateLimits []types.RateLimit
		numHits    uint64
		scanned    uint64
		nextKey    []byte
		scanCapHit bool
	)
	keyMode := len(pageReq.Key) != 0
	end := pageReq.Offset + pageReq.Limit

	for ; iterator.Valid(); iterator.Next() {
		if scanned >= maxFilteredQueryScanCap {
			nextKey = append([]byte(nil), iterator.Key()...)
			scanCapHit = true
			break
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

		accumulate := keyMode || (numHits >= pageReq.Offset && numHits < end)
		if accumulate {
			rateLimits = append(rateLimits, rateLimit)
		}
		numHits++

		if keyMode && uint64(len(rateLimits)) == pageReq.Limit {
			iterator.Next()
			if iterator.Valid() {
				nextKey = append([]byte(nil), iterator.Key()...)
			}
			break
		}
		if !keyMode && numHits == end+1 {
			if nextKey == nil {
				nextKey = append([]byte(nil), iterator.Key()...)
			}
			if !pageReq.CountTotal {
				break
			}
		}
	}

	res := &query.PageResponse{NextKey: nextKey}
	// Total is only meaningful for a full scan (not key-resumed, not cap-truncated).
	if pageReq.CountTotal && !keyMode && !scanCapHit {
		res.Total = numHits
	}
	return rateLimits, res, nil
}

func initFilteredPageRequest(p *query.PageRequest) *query.PageRequest {
	if p == nil {
		p = &query.PageRequest{}
	}
	pc := *p
	if len(pc.Key) == 0 {
		pc.Key = nil
	}
	if pc.Limit == 0 {
		pc.Limit = query.DefaultLimit
		pc.CountTotal = true
	}
	return &pc
}

func openFilteredIterator(store prefix.Store, start []byte, reverse bool) storetypes.Iterator {
	if !reverse {
		return store.Iterator(start, nil)
	}
	var end []byte
	if start != nil {
		itr := store.Iterator(start, nil)
		if itr.Valid() {
			itr.Next()
			end = itr.Key()
		}
		itr.Close()
	}
	return store.ReverseIterator(nil, end)
}

func (k Keeper) prefixedStore(ctx sdk.Context, keyPrefix []byte) prefix.Store {
	adapter := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	return prefix.NewStore(adapter, keyPrefix)
}
