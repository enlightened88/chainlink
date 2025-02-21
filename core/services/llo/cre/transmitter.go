package cre

import (
	"context"
	"errors"
	"strconv"

	"github.com/smartcontractkit/libocr/offchainreporting2/types"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"
	ocr2types "github.com/smartcontractkit/libocr/offchainreporting2plus/types"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	datastreamsllo "github.com/smartcontractkit/chainlink-data-streams/llo"

	capabilitiespb "github.com/smartcontractkit/chainlink-common/pkg/capabilities/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
	llotypes "github.com/smartcontractkit/chainlink-common/pkg/types/llo"
)

type Transmitter interface {
	llotypes.Transmitter
	services.Service
}

type TransmitterConfig struct {
	DonID uint32
	// TODO(@bolek):
	// You may add any kind of config you need here
}

var _ Transmitter = &transmitter{}

type transmitter struct {
	services.Service
	eng *services.Engine

	donID       uint32
	fromAccount ocr2types.Account

	// TODO(@bolek):
	// You may add anything you need here
}

func (c *TransmitterConfig) NewTransmitter(lggr logger.Logger) Transmitter {
	return c.newTransmitter(lggr)
}

func (c *TransmitterConfig) newTransmitter(lggr logger.Logger) *transmitter {
	t := &transmitter{
		donID:       c.DonID,
		fromAccount: ocr2types.Account(lggr.Name() + strconv.FormatUint(uint64(c.DonID), 10)),
	}

	t.Service, t.eng = services.Config{
		Name: "CRETransmitter",
		// TODO(@bolek):
		// You may add optional start, close, subservices hooks etc for your
		// trigger service here
	}.NewServiceEngine(lggr)

	return t
}

func (t *transmitter) FromAccount(context.Context) (ocr2types.Account, error) {
	return t.fromAccount, nil
}

func (t *transmitter) Transmit(
	ctx context.Context,
	cd ocr2types.ConfigDigest,
	seqNr uint64,
	report ocr3types.ReportWithInfo[llotypes.ReportInfo],
	sigs []types.AttributedOnchainSignature,
) error {
	switch report.Info.ReportFormat {
	case llotypes.ReportFormatCapabilityTrigger:
	default:
		// NOTE: Silently ignore non-capability format reports here. All
		// channels are broadcast to all transmitters but this transmitter only
		// cares about channels of type ReportFormatCapabilityTrigger
		return nil
	}
	switch report.Info.LifeCycleStage {
	case datastreamsllo.LifeCycleStageProduction:
	default:
		// NOTE: Ignore retirement and staging reports; for now we assume that
		// we only care about sending production reports.
		//
		// Support could be added in future e.g. for verifying blue-green
		// deploys etc.
		return nil
	}
	pbSigs := make([]*capabilitiespb.OCRAttributedOnchainSignature, len(sigs))
	for i, sig := range sigs {
		pbSigs[i] = &capabilitiespb.OCRAttributedOnchainSignature{
			Signer:    uint32(sig.Signer),
			Signature: sig.Signature,
		}
	}
	ev := &capabilitiespb.OCRTriggerEvent{
		ConfigDigest: cd[:],
		SeqNr:        seqNr,
		Report:       report.Report,
		Sigs:         pbSigs,
	}
	return t.processNewEvent(ctx, ev)
}

func (t *transmitter) processNewEvent(ctx context.Context, event *capabilitiespb.OCRTriggerEvent) error {
	// TODO(@bolek):
	// Implement the logic to process the new event here
	return errors.New("not implemented")
}
