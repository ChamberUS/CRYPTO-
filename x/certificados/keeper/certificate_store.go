package keeper

import (
	"context"
	"errors"
	"strings"
	"time"

	"cosmossdk.io/collections"

	"github.com/buynnex-corp/byx/x/certificados/types"
)

func (k Keeper) nextCertificateID(ctx context.Context) (uint64, error) {
	count, err := k.CertificateCount.Get(ctx)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			count = 0
		} else {
			return 0, err
		}
	}
	next := count + 1
	if err := k.CertificateCount.Set(ctx, next); err != nil {
		return 0, err
	}
	return next, nil
}

func (k Keeper) storeCertificate(ctx context.Context, cert types.Certificate) error {
	if err := k.Certificates.Set(ctx, cert.Id, cert); err != nil {
		return err
	}

	if err := k.ByOwner.Set(ctx, collections.Join(cert.Owner, cert.Id)); err != nil {
		return err
	}
	if err := k.ByMerchant.Set(ctx, collections.Join(cert.MerchantId, cert.Id)); err != nil {
		return err
	}
	serial := strings.ToLower(cert.SerialHash)
	if err := k.BySerialHash.Set(ctx, collections.Join(serial, cert.Id)); err != nil {
		return err
	}
	return nil
}

func (k Keeper) updateOwnerIndex(ctx context.Context, cert types.Certificate, newOwner string) error {
	if err := k.ByOwner.Remove(ctx, collections.Join(cert.Owner, cert.Id)); err != nil && !errors.Is(err, collections.ErrNotFound) {
		return err
	}
	cert.Owner = newOwner
	if err := k.Certificates.Set(ctx, cert.Id, cert); err != nil {
		return err
	}
	return k.ByOwner.Set(ctx, collections.Join(cert.Owner, cert.Id))
}

func (k Keeper) addServiceRecord(ctx context.Context, certID uint64, addedBy, kind, details string, at time.Time) (types.ServiceRecord, error) {
	count, err := k.ServiceRecordsCount.Get(ctx, certID)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			count = 0
		} else {
			return types.ServiceRecord{}, err
		}
	}
	next := count + 1
	record := types.ServiceRecord{
		CertificateId: certID,
		Index:         next,
		AddedBy:       addedBy,
		Kind:          kind,
		Details:       details,
		At:            at.UTC().Format(time.RFC3339),
	}
	if err := k.ServiceRecords.Set(ctx, collections.Join(certID, next), record); err != nil {
		return types.ServiceRecord{}, err
	}
	if err := k.ServiceRecordsCount.Set(ctx, certID, next); err != nil {
		return types.ServiceRecord{}, err
	}
	return record, nil
}

func (k Keeper) getCertificate(ctx context.Context, id uint64) (types.Certificate, error) {
	cert, err := k.Certificates.Get(ctx, id)
	if err != nil {
		if errors.Is(err, collections.ErrNotFound) {
			return types.Certificate{}, types.ErrCertificateNotFound
		}
		return types.Certificate{}, err
	}
	return cert, nil
}

func (k Keeper) ensureParams(ctx context.Context) types.Params {
	params, err := k.ParamsStore.Get(ctx)
	if err != nil {
		return types.DefaultParams()
	}
	return params
}

func (k Keeper) ensureEnabled(ctx context.Context) error {
	params := k.ensureParams(ctx)
	if !params.Enabled {
		return types.ErrModuleDisabled
	}
	return nil
}

func (k Keeper) ensureTransferAllowed(ctx context.Context) error {
	params := k.ensureParams(ctx)
	if !params.AllowTransfer {
		return types.ErrTransferNotAllowed
	}
	return nil
}
