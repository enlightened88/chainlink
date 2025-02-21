package cre

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/protobuf/proto"

	capabilitiespb "github.com/smartcontractkit/chainlink-common/pkg/capabilities/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	llotypes "github.com/smartcontractkit/chainlink-common/pkg/types/llo"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
	datastreamsllo "github.com/smartcontractkit/chainlink-data-streams/llo"
)

var _ datastreamsllo.ReportCodec = ReportCodecCapabilityTrigger{}

type ReportCodecCapabilityTrigger struct {
	lggr  logger.Logger
	donID uint32
}

func NewReportCodecCapabilityTrigger(lggr logger.Logger, donID uint32) ReportCodecCapabilityTrigger {
	return ReportCodecCapabilityTrigger{lggr, donID}
}

func (r ReportCodecCapabilityTrigger) Encode(ctx context.Context, report datastreamsllo.Report, cd llotypes.ChannelDefinition) ([]byte, error) {
	if len(cd.Streams) != len(report.Values) {
		// Invariant violation
		return nil, fmt.Errorf("capability trigger expected %d streams, got %d", len(cd.Streams), len(report.Values))
	}
	if report.Specimen {
		// Not supported for now
		return nil, errors.New("capability trigger encoder does not currently support specimen reports")
	}
	payload := make([]*StreamDecimal, len(report.Values))
	for i, stream := range report.Values {
		var d []byte
		switch stream.(type) {
		case nil:
			// Missing observations are ignored
			continue
		case *datastreamsllo.Decimal:
			var err error
			d, err = stream.MarshalBinary()
			if err != nil {
				return nil, fmt.Errorf("failed to marshal decimal: %w", err)
			}
		default:
			return nil, fmt.Errorf("only decimal StreamValues are supported, got: %T", stream)
		}
		payload[i] = &StreamDecimal{
			StreamID: cd.Streams[i].StreamID,
			Decimal:  d,
		}
	}
	ste := StreamsTriggerEvent{
		Payload:              payload,
		ObservationTimestamp: int64(report.ObservationTimestampSeconds),
	}
	outputs, err := values.WrapMap(ste)
	if err != nil {
		return nil, fmt.Errorf("failed to wrap map: %w", err)
	}
	p := &capabilitiespb.OCRTriggerReport{
		EventID:   r.eventID(report),
		Timestamp: int64(report.ObservationTimestampSeconds),
		Outputs:   values.ProtoMap(outputs),
	}

	b, err := proto.Marshal(p)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal capability trigger report: %w", err)
	}
	return b, nil
}

// eventID is expected to uniquely identify a (don, round)
func (r ReportCodecCapabilityTrigger) eventID(report datastreamsllo.Report) string {
	return fmt.Sprintf("streams_%d_%d", r.donID, report.ObservationTimestampSeconds)
}

type StreamsTriggerEvent struct {
	Payload              []*StreamDecimal
	ObservationTimestamp int64
}

type StreamDecimal struct {
	StreamID uint32
	Decimal  []byte
	// future: may add aggregation type {MODE, MEDIAN, etc...}
}
