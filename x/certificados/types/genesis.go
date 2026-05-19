package types

import (
	"strconv"
	"strings"

	errorsmod "cosmossdk.io/errors"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// DefaultGenesis returns the default genesis state.
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:           DefaultParams(),
		Certificates:     []Certificate{},
		CertificateCount: 0,
		ServiceRecords:   []ServiceRecord{},
		ServiceCounts:    []ServiceCount{},
	}
}

// Validate performs basic genesis state validation returning an error upon any failure.
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	seenCert := make(map[uint64]struct{}, len(gs.Certificates))
	var maxID uint64
	for _, c := range gs.Certificates {
		if _, ok := seenCert[c.Id]; ok {
			return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "duplicate certificate id %d", c.Id)
		}
		seenCert[c.Id] = struct{}{}
		if c.Id > maxID {
			maxID = c.Id
		}
		if strings.TrimSpace(c.SerialHash) == "" {
			return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "certificate serial_hash cannot be empty")
		}
	}

	if gs.CertificateCount < maxID {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "certificate_count %d < max certificate id %d", gs.CertificateCount, maxID)
	}

	seenService := make(map[string]struct{}, len(gs.ServiceRecords))
	for _, r := range gs.ServiceRecords {
		key := strings.Join([]string{strconv.FormatUint(r.CertificateId, 10), strconv.FormatUint(r.Index, 10)}, "/")
		if _, ok := seenService[key]; ok {
			return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "duplicate service record key "+key)
		}
		seenService[key] = struct{}{}
	}

	return nil
}
