// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: packetforward/v1/genesis.proto

package types

import (
	fmt "fmt"
	_ "github.com/cosmos/cosmos-sdk/types/tx/amino"
	_ "github.com/cosmos/gogoproto/gogoproto"
	proto "github.com/cosmos/gogoproto/proto"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

// GenesisState defines the packetforward genesis state
type GenesisState struct {
	// key - information about forwarded packet: src_channel
	// (parsedReceiver.Channel), src_port (parsedReceiver.Port), sequence value -
	// information about original packet for refunding if necessary: retries,
	// srcPacketSender, srcPacket.DestinationChannel, srcPacket.DestinationPort
	InFlightPackets map[string]InFlightPacket `protobuf:"bytes,2,rep,name=in_flight_packets,json=inFlightPackets,proto3" json:"in_flight_packets" yaml:"in_flight_packets" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (m *GenesisState) Reset()         { *m = GenesisState{} }
func (m *GenesisState) String() string { return proto.CompactTextString(m) }
func (*GenesisState) ProtoMessage()    {}
func (*GenesisState) Descriptor() ([]byte, []int) {
	return fileDescriptor_afd4e56ea31af982, []int{0}
}
func (m *GenesisState) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *GenesisState) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_GenesisState.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *GenesisState) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GenesisState.Merge(m, src)
}
func (m *GenesisState) XXX_Size() int {
	return m.Size()
}
func (m *GenesisState) XXX_DiscardUnknown() {
	xxx_messageInfo_GenesisState.DiscardUnknown(m)
}

var xxx_messageInfo_GenesisState proto.InternalMessageInfo

func (m *GenesisState) GetInFlightPackets() map[string]InFlightPacket {
	if m != nil {
		return m.InFlightPackets
	}
	return nil
}

// InFlightPacket contains information about original packet for
// writing the acknowledgement and refunding if necessary.
type InFlightPacket struct {
	OriginalSenderAddress  string `protobuf:"bytes,1,opt,name=original_sender_address,json=originalSenderAddress,proto3" json:"original_sender_address,omitempty"`
	RefundChannelId        string `protobuf:"bytes,2,opt,name=refund_channel_id,json=refundChannelId,proto3" json:"refund_channel_id,omitempty"`
	RefundPortId           string `protobuf:"bytes,3,opt,name=refund_port_id,json=refundPortId,proto3" json:"refund_port_id,omitempty"`
	PacketSrcChannelId     string `protobuf:"bytes,4,opt,name=packet_src_channel_id,json=packetSrcChannelId,proto3" json:"packet_src_channel_id,omitempty"`
	PacketSrcPortId        string `protobuf:"bytes,5,opt,name=packet_src_port_id,json=packetSrcPortId,proto3" json:"packet_src_port_id,omitempty"`
	PacketTimeoutTimestamp uint64 `protobuf:"varint,6,opt,name=packet_timeout_timestamp,json=packetTimeoutTimestamp,proto3" json:"packet_timeout_timestamp,omitempty"`
	PacketTimeoutHeight    string `protobuf:"bytes,7,opt,name=packet_timeout_height,json=packetTimeoutHeight,proto3" json:"packet_timeout_height,omitempty"`
	PacketData             []byte `protobuf:"bytes,8,opt,name=packet_data,json=packetData,proto3" json:"packet_data,omitempty"`
	RefundSequence         uint64 `protobuf:"varint,9,opt,name=refund_sequence,json=refundSequence,proto3" json:"refund_sequence,omitempty"`
	RetriesRemaining       int32  `protobuf:"varint,10,opt,name=retries_remaining,json=retriesRemaining,proto3" json:"retries_remaining,omitempty"`
	Timeout                uint64 `protobuf:"varint,11,opt,name=timeout,proto3" json:"timeout,omitempty"`
	Nonrefundable          bool   `protobuf:"varint,12,opt,name=nonrefundable,proto3" json:"nonrefundable,omitempty"`
}

func (m *InFlightPacket) Reset()         { *m = InFlightPacket{} }
func (m *InFlightPacket) String() string { return proto.CompactTextString(m) }
func (*InFlightPacket) ProtoMessage()    {}
func (*InFlightPacket) Descriptor() ([]byte, []int) {
	return fileDescriptor_afd4e56ea31af982, []int{1}
}
func (m *InFlightPacket) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *InFlightPacket) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_InFlightPacket.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *InFlightPacket) XXX_Merge(src proto.Message) {
	xxx_messageInfo_InFlightPacket.Merge(m, src)
}
func (m *InFlightPacket) XXX_Size() int {
	return m.Size()
}
func (m *InFlightPacket) XXX_DiscardUnknown() {
	xxx_messageInfo_InFlightPacket.DiscardUnknown(m)
}

var xxx_messageInfo_InFlightPacket proto.InternalMessageInfo

func (m *InFlightPacket) GetOriginalSenderAddress() string {
	if m != nil {
		return m.OriginalSenderAddress
	}
	return ""
}

func (m *InFlightPacket) GetRefundChannelId() string {
	if m != nil {
		return m.RefundChannelId
	}
	return ""
}

func (m *InFlightPacket) GetRefundPortId() string {
	if m != nil {
		return m.RefundPortId
	}
	return ""
}

func (m *InFlightPacket) GetPacketSrcChannelId() string {
	if m != nil {
		return m.PacketSrcChannelId
	}
	return ""
}

func (m *InFlightPacket) GetPacketSrcPortId() string {
	if m != nil {
		return m.PacketSrcPortId
	}
	return ""
}

func (m *InFlightPacket) GetPacketTimeoutTimestamp() uint64 {
	if m != nil {
		return m.PacketTimeoutTimestamp
	}
	return 0
}

func (m *InFlightPacket) GetPacketTimeoutHeight() string {
	if m != nil {
		return m.PacketTimeoutHeight
	}
	return ""
}

func (m *InFlightPacket) GetPacketData() []byte {
	if m != nil {
		return m.PacketData
	}
	return nil
}

func (m *InFlightPacket) GetRefundSequence() uint64 {
	if m != nil {
		return m.RefundSequence
	}
	return 0
}

func (m *InFlightPacket) GetRetriesRemaining() int32 {
	if m != nil {
		return m.RetriesRemaining
	}
	return 0
}

func (m *InFlightPacket) GetTimeout() uint64 {
	if m != nil {
		return m.Timeout
	}
	return 0
}

func (m *InFlightPacket) GetNonrefundable() bool {
	if m != nil {
		return m.Nonrefundable
	}
	return false
}

func init() {
	proto.RegisterType((*GenesisState)(nil), "packetforward.v1.GenesisState")
	proto.RegisterMapType((map[string]InFlightPacket)(nil), "packetforward.v1.GenesisState.InFlightPacketsEntry")
	proto.RegisterType((*InFlightPacket)(nil), "packetforward.v1.InFlightPacket")
}

func init() { proto.RegisterFile("packetforward/v1/genesis.proto", fileDescriptor_afd4e56ea31af982) }

var fileDescriptor_afd4e56ea31af982 = []byte{
	// 574 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x6c, 0x93, 0xcb, 0x4e, 0xdc, 0x3e,
	0x14, 0xc6, 0xc7, 0xdc, 0xf1, 0xcc, 0x9f, 0x8b, 0xff, 0xd0, 0x5a, 0x2c, 0x42, 0x34, 0x42, 0x6a,
	0x54, 0x44, 0x22, 0x40, 0x42, 0x88, 0x5d, 0xe9, 0x95, 0x1d, 0xca, 0xa0, 0x2e, 0xba, 0x89, 0x3c,
	0xb1, 0xc9, 0x58, 0x24, 0x76, 0x6a, 0x7b, 0x06, 0xcd, 0xb2, 0x6f, 0xd0, 0x37, 0xe8, 0xeb, 0xb0,
	0x64, 0xd9, 0x15, 0xaa, 0xe0, 0x0d, 0xba, 0xea, 0xb2, 0x1a, 0x3b, 0xa1, 0x93, 0xd2, 0x4d, 0xe2,
	0x9c, 0xdf, 0xf7, 0x7d, 0x3e, 0x3e, 0x8a, 0xa1, 0x57, 0x92, 0xf4, 0x8a, 0x99, 0x4b, 0xa9, 0xae,
	0x89, 0xa2, 0xd1, 0x68, 0x3f, 0xca, 0x98, 0x60, 0x9a, 0xeb, 0xb0, 0x54, 0xd2, 0x48, 0xb4, 0xd6,
	0xe0, 0xe1, 0x68, 0x7f, 0x6b, 0x23, 0x93, 0x99, 0xb4, 0x30, 0x9a, 0xac, 0x9c, 0x6e, 0x6b, 0x9d,
	0x14, 0x5c, 0xc8, 0xc8, 0x3e, 0x5d, 0xa9, 0xfb, 0x0b, 0xc0, 0xce, 0x7b, 0x17, 0xd6, 0x33, 0xc4,
	0x30, 0xf4, 0x05, 0xc0, 0x75, 0x2e, 0x92, 0xcb, 0x9c, 0x67, 0x03, 0x93, 0xb8, 0x60, 0x8d, 0x67,
	0xfc, 0xd9, 0xa0, 0x7d, 0x70, 0x18, 0xfe, 0xbd, 0x51, 0x38, 0xed, 0x0d, 0xcf, 0xc4, 0x3b, 0x6b,
	0x3b, 0x77, 0xae, 0xb7, 0xc2, 0xa8, 0xf1, 0xa9, 0x7f, 0x73, 0xb7, 0xdd, 0xfa, 0x79, 0xb7, 0x8d,
	0xc7, 0xa4, 0xc8, 0x4f, 0xba, 0x4f, 0xb2, 0xbb, 0xf1, 0x2a, 0x6f, 0xfa, 0xb6, 0x28, 0xdc, 0xf8,
	0x57, 0x14, 0x5a, 0x83, 0xb3, 0x57, 0x6c, 0x8c, 0x81, 0x0f, 0x82, 0xe5, 0x78, 0xb2, 0x44, 0x47,
	0x70, 0x7e, 0x44, 0xf2, 0x21, 0xc3, 0x33, 0x3e, 0x08, 0xda, 0x07, 0xfe, 0xd3, 0x06, 0x9b, 0x41,
	0xb1, 0x93, 0x9f, 0xcc, 0x1c, 0x83, 0xee, 0xb7, 0x39, 0xb8, 0xd2, 0xa4, 0xe8, 0x08, 0x3e, 0x97,
	0x8a, 0x67, 0x5c, 0x90, 0x3c, 0xd1, 0x4c, 0x50, 0xa6, 0x12, 0x42, 0xa9, 0x62, 0x5a, 0x57, 0x9b,
	0x6e, 0xd6, 0xb8, 0x67, 0xe9, 0x2b, 0x07, 0xd1, 0x4b, 0xb8, 0xae, 0xd8, 0xe5, 0x50, 0xd0, 0x24,
	0x1d, 0x10, 0x21, 0x58, 0x9e, 0x70, 0x6a, 0x5b, 0x5a, 0x8e, 0x57, 0x1d, 0x78, 0xed, 0xea, 0x67,
	0x14, 0xed, 0xc0, 0x95, 0x4a, 0x5b, 0x4a, 0x65, 0x26, 0xc2, 0x59, 0x2b, 0xec, 0xb8, 0xea, 0xb9,
	0x54, 0xe6, 0x8c, 0xa2, 0x7d, 0xb8, 0xe9, 0x8e, 0x92, 0x68, 0x95, 0x4e, 0xa7, 0xce, 0x59, 0x31,
	0x72, 0xb0, 0xa7, 0xd2, 0x3f, 0xc1, 0xbb, 0x10, 0x4d, 0x59, 0xea, 0xf0, 0x79, 0xd7, 0xc5, 0xa3,
	0xbe, 0xca, 0x3f, 0x86, 0xb8, 0x12, 0x1b, 0x5e, 0x30, 0x39, 0x74, 0x6f, 0x6d, 0x48, 0x51, 0xe2,
	0x05, 0x1f, 0x04, 0x73, 0xf1, 0x33, 0xc7, 0x2f, 0x1c, 0xbe, 0xa8, 0x29, 0x3a, 0x78, 0xec, 0xac,
	0x76, 0x0e, 0xd8, 0x64, 0x84, 0x78, 0xd1, 0xee, 0xf4, 0x7f, 0xc3, 0xf6, 0xc1, 0x22, 0xb4, 0x0d,
	0xdb, 0x95, 0x87, 0x12, 0x43, 0xf0, 0x92, 0x0f, 0x82, 0x4e, 0x0c, 0x5d, 0xe9, 0x0d, 0x31, 0x04,
	0xbd, 0x80, 0xd5, 0x9c, 0x12, 0xcd, 0x3e, 0x0f, 0x99, 0x48, 0x19, 0x5e, 0xb6, 0x5d, 0x54, 0xb3,
	0xea, 0x55, 0x55, 0xb4, 0x3b, 0x99, 0xb4, 0x51, 0x9c, 0xe9, 0x44, 0xb1, 0x82, 0x70, 0xc1, 0x45,
	0x86, 0xa1, 0x0f, 0x82, 0xf9, 0x78, 0xad, 0x02, 0x71, 0x5d, 0x47, 0x18, 0x2e, 0x56, 0x3d, 0xe2,
	0xb6, 0x4d, 0xab, 0x3f, 0xd1, 0x0e, 0xfc, 0x4f, 0x48, 0xe1, 0xb2, 0x49, 0x3f, 0x67, 0xb8, 0xe3,
	0x83, 0x60, 0x29, 0x6e, 0x16, 0x4f, 0xcb, 0x9b, 0x7b, 0x0f, 0xdc, 0xde, 0x7b, 0xe0, 0xc7, 0xbd,
	0x07, 0xbe, 0x3e, 0x78, 0xad, 0xdb, 0x07, 0xaf, 0xf5, 0xfd, 0xc1, 0x6b, 0x7d, 0xfa, 0x98, 0x71,
	0x33, 0x18, 0xf6, 0xc3, 0x54, 0x16, 0x51, 0x2a, 0x75, 0x21, 0x75, 0xc4, 0xfb, 0xe9, 0x1e, 0x29,
	0x4b, 0x1d, 0x15, 0x9c, 0xd2, 0x9c, 0x5d, 0x13, 0xc5, 0x22, 0x77, 0xc2, 0xbd, 0xea, 0x77, 0xdc,
	0x9b, 0x22, 0xa3, 0xe3, 0xa8, 0x79, 0xa9, 0xcd, 0xb8, 0x64, 0xba, 0xbf, 0x60, 0x6f, 0xe5, 0xe1,
	0xef, 0x00, 0x00, 0x00, 0xff, 0xff, 0xa5, 0x8e, 0x69, 0x25, 0xf2, 0x03, 0x00, 0x00,
}

func (m *GenesisState) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *GenesisState) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *GenesisState) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.InFlightPackets) > 0 {
		for k := range m.InFlightPackets {
			v := m.InFlightPackets[k]
			baseI := i
			{
				size, err := (&v).MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintGenesis(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x12
			i -= len(k)
			copy(dAtA[i:], k)
			i = encodeVarintGenesis(dAtA, i, uint64(len(k)))
			i--
			dAtA[i] = 0xa
			i = encodeVarintGenesis(dAtA, i, uint64(baseI-i))
			i--
			dAtA[i] = 0x12
		}
	}
	return len(dAtA) - i, nil
}

func (m *InFlightPacket) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *InFlightPacket) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *InFlightPacket) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Nonrefundable {
		i--
		if m.Nonrefundable {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x60
	}
	if m.Timeout != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.Timeout))
		i--
		dAtA[i] = 0x58
	}
	if m.RetriesRemaining != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.RetriesRemaining))
		i--
		dAtA[i] = 0x50
	}
	if m.RefundSequence != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.RefundSequence))
		i--
		dAtA[i] = 0x48
	}
	if len(m.PacketData) > 0 {
		i -= len(m.PacketData)
		copy(dAtA[i:], m.PacketData)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.PacketData)))
		i--
		dAtA[i] = 0x42
	}
	if len(m.PacketTimeoutHeight) > 0 {
		i -= len(m.PacketTimeoutHeight)
		copy(dAtA[i:], m.PacketTimeoutHeight)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.PacketTimeoutHeight)))
		i--
		dAtA[i] = 0x3a
	}
	if m.PacketTimeoutTimestamp != 0 {
		i = encodeVarintGenesis(dAtA, i, uint64(m.PacketTimeoutTimestamp))
		i--
		dAtA[i] = 0x30
	}
	if len(m.PacketSrcPortId) > 0 {
		i -= len(m.PacketSrcPortId)
		copy(dAtA[i:], m.PacketSrcPortId)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.PacketSrcPortId)))
		i--
		dAtA[i] = 0x2a
	}
	if len(m.PacketSrcChannelId) > 0 {
		i -= len(m.PacketSrcChannelId)
		copy(dAtA[i:], m.PacketSrcChannelId)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.PacketSrcChannelId)))
		i--
		dAtA[i] = 0x22
	}
	if len(m.RefundPortId) > 0 {
		i -= len(m.RefundPortId)
		copy(dAtA[i:], m.RefundPortId)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.RefundPortId)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.RefundChannelId) > 0 {
		i -= len(m.RefundChannelId)
		copy(dAtA[i:], m.RefundChannelId)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.RefundChannelId)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.OriginalSenderAddress) > 0 {
		i -= len(m.OriginalSenderAddress)
		copy(dAtA[i:], m.OriginalSenderAddress)
		i = encodeVarintGenesis(dAtA, i, uint64(len(m.OriginalSenderAddress)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintGenesis(dAtA []byte, offset int, v uint64) int {
	offset -= sovGenesis(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *GenesisState) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.InFlightPackets) > 0 {
		for k, v := range m.InFlightPackets {
			_ = k
			_ = v
			l = v.Size()
			mapEntrySize := 1 + len(k) + sovGenesis(uint64(len(k))) + 1 + l + sovGenesis(uint64(l))
			n += mapEntrySize + 1 + sovGenesis(uint64(mapEntrySize))
		}
	}
	return n
}

func (m *InFlightPacket) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.OriginalSenderAddress)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	l = len(m.RefundChannelId)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	l = len(m.RefundPortId)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	l = len(m.PacketSrcChannelId)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	l = len(m.PacketSrcPortId)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	if m.PacketTimeoutTimestamp != 0 {
		n += 1 + sovGenesis(uint64(m.PacketTimeoutTimestamp))
	}
	l = len(m.PacketTimeoutHeight)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	l = len(m.PacketData)
	if l > 0 {
		n += 1 + l + sovGenesis(uint64(l))
	}
	if m.RefundSequence != 0 {
		n += 1 + sovGenesis(uint64(m.RefundSequence))
	}
	if m.RetriesRemaining != 0 {
		n += 1 + sovGenesis(uint64(m.RetriesRemaining))
	}
	if m.Timeout != 0 {
		n += 1 + sovGenesis(uint64(m.Timeout))
	}
	if m.Nonrefundable {
		n += 2
	}
	return n
}

func sovGenesis(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozGenesis(x uint64) (n int) {
	return sovGenesis(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *GenesisState) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGenesis
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: GenesisState: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: GenesisState: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field InFlightPackets", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.InFlightPackets == nil {
				m.InFlightPackets = make(map[string]InFlightPacket)
			}
			var mapkey string
			mapvalue := &InFlightPacket{}
			for iNdEx < postIndex {
				entryPreIndex := iNdEx
				var wire uint64
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowGenesis
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					wire |= uint64(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				fieldNum := int32(wire >> 3)
				if fieldNum == 1 {
					var stringLenmapkey uint64
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowGenesis
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						stringLenmapkey |= uint64(b&0x7F) << shift
						if b < 0x80 {
							break
						}
					}
					intStringLenmapkey := int(stringLenmapkey)
					if intStringLenmapkey < 0 {
						return ErrInvalidLengthGenesis
					}
					postStringIndexmapkey := iNdEx + intStringLenmapkey
					if postStringIndexmapkey < 0 {
						return ErrInvalidLengthGenesis
					}
					if postStringIndexmapkey > l {
						return io.ErrUnexpectedEOF
					}
					mapkey = string(dAtA[iNdEx:postStringIndexmapkey])
					iNdEx = postStringIndexmapkey
				} else if fieldNum == 2 {
					var mapmsglen int
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowGenesis
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						mapmsglen |= int(b&0x7F) << shift
						if b < 0x80 {
							break
						}
					}
					if mapmsglen < 0 {
						return ErrInvalidLengthGenesis
					}
					postmsgIndex := iNdEx + mapmsglen
					if postmsgIndex < 0 {
						return ErrInvalidLengthGenesis
					}
					if postmsgIndex > l {
						return io.ErrUnexpectedEOF
					}
					mapvalue = &InFlightPacket{}
					if err := mapvalue.Unmarshal(dAtA[iNdEx:postmsgIndex]); err != nil {
						return err
					}
					iNdEx = postmsgIndex
				} else {
					iNdEx = entryPreIndex
					skippy, err := skipGenesis(dAtA[iNdEx:])
					if err != nil {
						return err
					}
					if (skippy < 0) || (iNdEx+skippy) < 0 {
						return ErrInvalidLengthGenesis
					}
					if (iNdEx + skippy) > postIndex {
						return io.ErrUnexpectedEOF
					}
					iNdEx += skippy
				}
			}
			m.InFlightPackets[mapkey] = *mapvalue
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipGenesis(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthGenesis
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *InFlightPacket) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowGenesis
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: InFlightPacket: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: InFlightPacket: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field OriginalSenderAddress", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.OriginalSenderAddress = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field RefundChannelId", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.RefundChannelId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field RefundPortId", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.RefundPortId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PacketSrcChannelId", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.PacketSrcChannelId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PacketSrcPortId", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.PacketSrcPortId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 6:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field PacketTimeoutTimestamp", wireType)
			}
			m.PacketTimeoutTimestamp = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.PacketTimeoutTimestamp |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 7:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PacketTimeoutHeight", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.PacketTimeoutHeight = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 8:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PacketData", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthGenesis
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthGenesis
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.PacketData = append(m.PacketData[:0], dAtA[iNdEx:postIndex]...)
			if m.PacketData == nil {
				m.PacketData = []byte{}
			}
			iNdEx = postIndex
		case 9:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field RefundSequence", wireType)
			}
			m.RefundSequence = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.RefundSequence |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 10:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field RetriesRemaining", wireType)
			}
			m.RetriesRemaining = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.RetriesRemaining |= int32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 11:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Timeout", wireType)
			}
			m.Timeout = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Timeout |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 12:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Nonrefundable", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.Nonrefundable = bool(v != 0)
		default:
			iNdEx = preIndex
			skippy, err := skipGenesis(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthGenesis
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipGenesis(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowGenesis
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowGenesis
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthGenesis
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupGenesis
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthGenesis
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthGenesis        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowGenesis          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupGenesis = fmt.Errorf("proto: unexpected end of group")
)
