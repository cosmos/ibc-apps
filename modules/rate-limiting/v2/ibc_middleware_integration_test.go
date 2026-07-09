package v2_test

import (
	"encoding/hex"
	"errors"
	"testing"
	"time"

	"github.com/cosmos/gogoproto/proto"
	"github.com/cosmos/ibc-apps/modules/rate-limiting/v10/testing/simapp"
	ratelimittypes "github.com/cosmos/ibc-apps/modules/rate-limiting/v10/types"
	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	abci "github.com/cometbft/cometbft/abci/types"

	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v10/modules/core/04-channel/types"
	channeltypesv2 "github.com/cosmos/ibc-go/v10/modules/core/04-channel/v2/types"
	hostv2 "github.com/cosmos/ibc-go/v10/modules/core/24-host/v2"
	ibctesting "github.com/cosmos/ibc-go/v10/testing"
)

const rateLimitChannelValue = int64(1000)

type RateLimitMiddlewareTestSuite struct {
	suite.Suite

	coordinator *ibctesting.Coordinator
	chainA      *ibctesting.TestChain
	chainB      *ibctesting.TestChain
	path        *ibctesting.Path
}

func TestRateLimitMiddlewareTestSuite(t *testing.T) {
	suite.Run(t, new(RateLimitMiddlewareTestSuite))
}

func (s *RateLimitMiddlewareTestSuite) SetupTest() {
	ibctesting.DefaultTestingAppInit = simapp.SetupTestingApp
	s.coordinator = ibctesting.NewCoordinator(s.T(), 2)
	s.chainA = s.coordinator.GetChain(ibctesting.GetChainID(1))
	s.chainB = s.coordinator.GetChain(ibctesting.GetChainID(2))
	s.path = ibctesting.NewPath(s.chainA, s.chainB)
	s.path.SetupV2()

	s.seedHourEpoch(s.chainA)
	s.seedHourEpoch(s.chainB)
}

// seedHourEpoch aligns the hour epoch with the current block time. The ibctesting
// package initializes genesis with a zero block time, which leaves the epoch start
// time at the zero value and causes the BeginBlocker to reset all rate limit flows
// on every block.
func (s *RateLimitMiddlewareTestSuite) seedHourEpoch(chain *ibctesting.TestChain) {
	ctx := chain.GetContext()
	simapp.GetSimApp(chain).RatelimitKeeper.SetHourEpoch(ctx, ratelimittypes.HourEpoch{
		EpochNumber:      uint64(ctx.BlockTime().Hour()), //nolint:gosec
		Duration:         time.Hour,
		EpochStartTime:   ctx.BlockTime().Truncate(time.Hour),
		EpochStartHeight: ctx.BlockHeight(),
	})
}

func (s *RateLimitMiddlewareTestSuite) TestV2TransferSuccessUpdatesFlows() {
	amount := sdkmath.NewInt(10)
	recvDenom := s.voucherDenom()

	s.setRateLimit(s.chainA, sdk.DefaultBondDenom, s.path.EndpointA.ClientID, 100, 100)
	s.setRateLimit(s.chainB, recvDenom, s.path.EndpointB.ClientID, 100, 100)

	senderInitialBalance := s.balance(s.chainA, s.chainA.SenderAccount.GetAddress(), sdk.DefaultBondDenom)
	payload := s.transferPayload(amount)

	packet, err := s.path.EndpointA.MsgSendPacket(s.timeoutTimestamp(time.Hour), payload)
	s.Require().NoError(err)

	s.assertFlow(s.chainA, sdk.DefaultBondDenom, s.path.EndpointA.ClientID, sdkmath.ZeroInt(), amount)
	s.assertPendingPacket(s.chainA, sdk.DefaultBondDenom, s.path.EndpointA.ClientID, packet.Sequence, true)

	ack, err := s.msgRecvPacketWithAck(s.path.EndpointB, packet)
	s.Require().NoError(err)
	s.Require().Len(ack.AppAcknowledgements, 1)
	s.Require().Equal(channeltypes.NewResultAcknowledgement([]byte{byte(1)}).Acknowledgement(), ack.AppAcknowledgements[0])

	s.assertFlow(s.chainB, recvDenom, s.path.EndpointB.ClientID, amount, sdkmath.ZeroInt())
	s.Require().Equal(amount, s.balance(s.chainB, s.chainB.SenderAccount.GetAddress(), recvDenom).Amount)

	err = s.path.EndpointA.MsgAcknowledgePacket(packet, ack)
	s.Require().NoError(err)

	s.assertFlow(s.chainA, sdk.DefaultBondDenom, s.path.EndpointA.ClientID, sdkmath.ZeroInt(), amount)
	s.assertPendingPacket(s.chainA, sdk.DefaultBondDenom, s.path.EndpointA.ClientID, packet.Sequence, false)
	s.Require().Equal(senderInitialBalance.Amount.Sub(amount), s.balance(s.chainA, s.chainA.SenderAccount.GetAddress(), sdk.DefaultBondDenom).Amount)
}

func (s *RateLimitMiddlewareTestSuite) TestV2TransferSendDenied() {
	amount := sdkmath.NewInt(11)
	s.setRateLimit(s.chainA, sdk.DefaultBondDenom, s.path.EndpointA.ClientID, 1, 100)

	senderInitialBalance := s.balance(s.chainA, s.chainA.SenderAccount.GetAddress(), sdk.DefaultBondDenom)
	payload := s.transferPayload(amount)

	_, err := s.path.EndpointA.MsgSendPacket(s.timeoutTimestamp(time.Hour), payload)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), ratelimittypes.ErrQuotaExceeded.Error())

	s.assertFlow(s.chainA, sdk.DefaultBondDenom, s.path.EndpointA.ClientID, sdkmath.ZeroInt(), sdkmath.ZeroInt())
	s.assertPendingPacket(s.chainA, sdk.DefaultBondDenom, s.path.EndpointA.ClientID, 1, false)
	s.Require().Equal(senderInitialBalance, s.balance(s.chainA, s.chainA.SenderAccount.GetAddress(), sdk.DefaultBondDenom))
}

func (s *RateLimitMiddlewareTestSuite) TestV2TransferReceiveDeniedUndoSendOnErrorAck() {
	amount := sdkmath.NewInt(11)
	recvDenom := s.voucherDenom()

	s.setRateLimit(s.chainA, sdk.DefaultBondDenom, s.path.EndpointA.ClientID, 100, 100)
	s.setRateLimit(s.chainB, recvDenom, s.path.EndpointB.ClientID, 100, 1)

	senderInitialBalance := s.balance(s.chainA, s.chainA.SenderAccount.GetAddress(), sdk.DefaultBondDenom)
	payload := s.transferPayload(amount)

	packet, err := s.path.EndpointA.MsgSendPacket(s.timeoutTimestamp(time.Hour), payload)
	s.Require().NoError(err)
	s.assertFlow(s.chainA, sdk.DefaultBondDenom, s.path.EndpointA.ClientID, sdkmath.ZeroInt(), amount)
	s.assertPendingPacket(s.chainA, sdk.DefaultBondDenom, s.path.EndpointA.ClientID, packet.Sequence, true)

	ack, err := s.msgRecvPacketWithAck(s.path.EndpointB, packet)
	s.Require().NoError(err)
	s.Require().Len(ack.AppAcknowledgements, 1)
	s.Require().Equal(channeltypesv2.ErrorAcknowledgement[:], ack.AppAcknowledgements[0])
	s.assertFlow(s.chainB, recvDenom, s.path.EndpointB.ClientID, sdkmath.ZeroInt(), sdkmath.ZeroInt())
	s.Require().True(s.balance(s.chainB, s.chainB.SenderAccount.GetAddress(), recvDenom).IsZero())

	err = s.path.EndpointA.MsgAcknowledgePacket(packet, ack)
	s.Require().NoError(err)

	s.assertFlow(s.chainA, sdk.DefaultBondDenom, s.path.EndpointA.ClientID, sdkmath.ZeroInt(), sdkmath.ZeroInt())
	s.assertPendingPacket(s.chainA, sdk.DefaultBondDenom, s.path.EndpointA.ClientID, packet.Sequence, false)
	s.Require().Equal(senderInitialBalance, s.balance(s.chainA, s.chainA.SenderAccount.GetAddress(), sdk.DefaultBondDenom))
}

func (s *RateLimitMiddlewareTestSuite) TestV2TransferTimeoutUndoSend() {
	amount := sdkmath.NewInt(10)
	s.setRateLimit(s.chainA, sdk.DefaultBondDenom, s.path.EndpointA.ClientID, 100, 100)

	senderInitialBalance := s.balance(s.chainA, s.chainA.SenderAccount.GetAddress(), sdk.DefaultBondDenom)
	payload := s.transferPayload(amount)

	packet, err := s.path.EndpointA.MsgSendPacket(s.timeoutTimestamp(time.Second), payload)
	s.Require().NoError(err)
	s.assertFlow(s.chainA, sdk.DefaultBondDenom, s.path.EndpointA.ClientID, sdkmath.ZeroInt(), amount)
	s.assertPendingPacket(s.chainA, sdk.DefaultBondDenom, s.path.EndpointA.ClientID, packet.Sequence, true)
	s.Require().Equal(senderInitialBalance.Amount.Sub(amount), s.balance(s.chainA, s.chainA.SenderAccount.GetAddress(), sdk.DefaultBondDenom).Amount)

	s.Require().NoError(s.path.EndpointA.UpdateClient())
	err = s.path.EndpointA.MsgTimeoutPacket(packet)
	s.Require().NoError(err)

	s.assertFlow(s.chainA, sdk.DefaultBondDenom, s.path.EndpointA.ClientID, sdkmath.ZeroInt(), sdkmath.ZeroInt())
	s.assertPendingPacket(s.chainA, sdk.DefaultBondDenom, s.path.EndpointA.ClientID, packet.Sequence, false)
	s.Require().Equal(senderInitialBalance, s.balance(s.chainA, s.chainA.SenderAccount.GetAddress(), sdk.DefaultBondDenom))
}

// msgRecvPacketWithAck sends a MsgRecvPacket on the associated endpoint and returns
// the acknowledgement written as a result. It mirrors ibctesting's MsgRecvPacket,
// which discards the acknowledgement.
func (s *RateLimitMiddlewareTestSuite) msgRecvPacketWithAck(endpoint *ibctesting.Endpoint, packet channeltypesv2.Packet) (channeltypesv2.Acknowledgement, error) {
	packetKey := hostv2.PacketCommitmentKey(packet.SourceClient, packet.Sequence)
	proof, proofHeight := endpoint.Counterparty.QueryProof(packetKey)

	msg := channeltypesv2.NewMsgRecvPacket(packet, proof, proofHeight, endpoint.Chain.SenderAccount.GetAddress().String())

	res, err := endpoint.Chain.SendMsgs(msg)
	if err != nil {
		return channeltypesv2.Acknowledgement{}, err
	}

	if err := endpoint.Counterparty.UpdateClient(); err != nil {
		return channeltypesv2.Acknowledgement{}, err
	}

	return parseAckFromEvents(res.Events)
}

func parseAckFromEvents(events []abci.Event) (channeltypesv2.Acknowledgement, error) {
	for _, event := range events {
		if event.Type != channeltypesv2.EventTypeWriteAck {
			continue
		}
		for _, attr := range event.Attributes {
			if attr.Key != channeltypesv2.AttributeKeyEncodedAckHex {
				continue
			}
			ackBz, err := hex.DecodeString(attr.Value)
			if err != nil {
				return channeltypesv2.Acknowledgement{}, err
			}
			var ack channeltypesv2.Acknowledgement
			if err := proto.Unmarshal(ackBz, &ack); err != nil {
				return channeltypesv2.Acknowledgement{}, err
			}
			return ack, nil
		}
	}
	return channeltypesv2.Acknowledgement{}, errors.New("acknowledgement event not found")
}

func (s *RateLimitMiddlewareTestSuite) setRateLimit(chain *ibctesting.TestChain, denom, clientID string, maxPercentSend, maxPercentRecv int64) {
	simapp.GetSimApp(chain).RatelimitKeeper.SetRateLimit(chain.GetContext(), ratelimittypes.RateLimit{
		Path: &ratelimittypes.Path{
			Denom:             denom,
			ChannelOrClientId: clientID,
		},
		Quota: &ratelimittypes.Quota{
			MaxPercentSend: sdkmath.NewInt(maxPercentSend),
			MaxPercentRecv: sdkmath.NewInt(maxPercentRecv),
			DurationHours:  1,
		},
		Flow: &ratelimittypes.Flow{
			Inflow:       sdkmath.ZeroInt(),
			Outflow:      sdkmath.ZeroInt(),
			ChannelValue: sdkmath.NewInt(rateLimitChannelValue),
		},
	})
}

func (s *RateLimitMiddlewareTestSuite) transferPayload(amount sdkmath.Int) channeltypesv2.Payload {
	packetData := transfertypes.NewFungibleTokenPacketData(
		sdk.DefaultBondDenom,
		amount.String(),
		s.chainA.SenderAccount.GetAddress().String(),
		s.chainB.SenderAccount.GetAddress().String(),
		"",
	)
	bz := s.chainA.Codec.MustMarshal(&packetData)

	return channeltypesv2.NewPayload(transfertypes.PortID, transfertypes.PortID, transfertypes.V1, transfertypes.EncodingProtobuf, bz)
}

func (s *RateLimitMiddlewareTestSuite) voucherDenom() string {
	return transfertypes.NewDenom(
		sdk.DefaultBondDenom,
		transfertypes.NewHop(transfertypes.PortID, s.path.EndpointB.ClientID),
	).IBCDenom()
}

func (s *RateLimitMiddlewareTestSuite) timeoutTimestamp(duration time.Duration) uint64 {
	return uint64(s.chainB.GetContext().BlockTime().Add(duration).Unix())
}

func (s *RateLimitMiddlewareTestSuite) assertFlow(chain *ibctesting.TestChain, denom, clientID string, expectedInflow, expectedOutflow sdkmath.Int) {
	rateLimit, found := simapp.GetSimApp(chain).RatelimitKeeper.GetRateLimit(chain.GetContext(), denom, clientID)
	s.Require().True(found)
	s.Require().True(rateLimit.Flow.Inflow.Equal(expectedInflow), "expected inflow %s, got %s", expectedInflow, rateLimit.Flow.Inflow)
	s.Require().True(rateLimit.Flow.Outflow.Equal(expectedOutflow), "expected outflow %s, got %s", expectedOutflow, rateLimit.Flow.Outflow)
}

func (s *RateLimitMiddlewareTestSuite) assertPendingPacket(chain *ibctesting.TestChain, denom, clientID string, sequence uint64, expected bool) {
	found, err := simapp.GetSimApp(chain).RatelimitKeeper.CheckPacketSentDuringCurrentQuota(chain.GetContext(), clientID, sequence, denom)
	s.Require().NoError(err)
	s.Require().Equal(expected, found)
}

func (s *RateLimitMiddlewareTestSuite) balance(chain *ibctesting.TestChain, address sdk.AccAddress, denom string) sdk.Coin {
	return simapp.GetSimApp(chain).BankKeeper.GetBalance(chain.GetContext(), address, denom)
}
