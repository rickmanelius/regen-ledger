package server

import (
	"bytes"
	"context"
	"fmt"
	"github.com/multiformats/go-multihash"

	gocid "github.com/ipfs/go-cid"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	gogotypes "github.com/gogo/protobuf/types"

	"github.com/regen-network/regen-ledger/x/data"
)

var _ data.MsgServer = serverImpl{}

func (s serverImpl) AnchorData(goCtx context.Context, request *data.MsgAnchorDataRequest) (*data.MsgAnchorDataResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	cidBz := request.Cid
	if s.anchorTable.Has(ctx, cidBz) {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, fmt.Sprintf("CID f%x is already anchored", cidBz))
	}

	err := s.anchorCid(ctx, cidBz)
	if err != nil {
		return nil, err
	}

	return &data.MsgAnchorDataResponse{}, nil
}

func (s serverImpl) anchorCidIfNeeded(ctx sdk.Context, cid []byte) error {
	if s.anchorTable.Has(ctx, cid) {
		return nil
	}

	return s.anchorCid(ctx, cid)
}

func (s serverImpl) anchorCid(ctx sdk.Context, cidBytes []byte) error {
	timestamp, err := gogotypes.TimestampProto(ctx.BlockTime())
	if err != nil {
		return sdkerrors.Wrap(err, "invalid block time")
	}

	err = s.anchorTable.Create(ctx, cidBytes, timestamp)
	if err != nil {
		return sdkerrors.Wrap(err, "error anchoring data")
	}

	return ctx.EventManager().EmitTypedEvent(&data.EventAnchorData{Cid: cidBytes})
}

func (s serverImpl) SignData(goCtx context.Context, request *data.MsgSignDataRequest) (*data.MsgSignDataResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	cidBz := request.Cid
	err := s.anchorCidIfNeeded(ctx, cidBz)
	if err != nil {
		return nil, err
	}

	// TODO: index both cid and signer in key
	var signers data.Signers
	if s.signersTable.Has(ctx, cidBz) {
		err = s.signersTable.GetOne(ctx, cidBz, &signers)
		if err != nil {
			return nil, err
		}

		// merge signers
		seen := map[string]bool{}
		for _, signer := range signers.Signers {
			seen[signer] = true
		}

		for _, signer := range request.Signers {
			_, found := seen[signer]
			if !found {
				signers.Signers = append(signers.Signers, signer)
			}
		}
	}

	err = s.signersTable.Save(ctx, cidBz, &signers)
	if err != nil {
		return nil, err
	}

	err = ctx.EventManager().EmitTypedEvent(&data.EventSignData{
		Cid:     cidBz,
		Signers: request.Signers,
	})
	if err != nil {
		return nil, err
	}

	return &data.MsgSignDataResponse{}, nil
}

func (s serverImpl) StoreData(goCtx context.Context, request *data.MsgStoreDataRequest) (*data.MsgStoreDataResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	cidBz := request.Cid
	err := s.anchorCidIfNeeded(ctx, cidBz)
	if err != nil {
		return nil, err
	}

	store := ctx.KVStore(s.storeKey)
	if store.Has(cidBz) {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, fmt.Sprintf("CID %s already has stored data", cidBz))
	}

	cid, err := gocid.Cast(cidBz)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, fmt.Sprintf("bad CID f%x", cidBz))
	}

	mh := cid.Hash()

	decodedMultihash, err := multihash.Decode(mh)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "can't retrieve multihash")
	}

	switch decodedMultihash.Name {
	case "sha2-256":
		// TODO: gas
	case "blake2b-256":
		// TODO: gas
	default:
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, fmt.Sprintf("unsupported hash function %s", decodedMultihash.Name))
	}

	reqMh, err := multihash.Sum(request.Content, decodedMultihash.Code, -1)
	if err != nil {
		return nil, sdkerrors.Wrap(err, fmt.Sprintf("unable to perform multihash"))
	}

	if !bytes.Equal(mh, reqMh) {
		return nil, sdkerrors.Wrap(err, fmt.Sprintf("unable to perform multihash"))
	}

	store.Set(cidBz, request.Content)

	err = ctx.EventManager().EmitTypedEvent(&data.EventStoreData{Cid: cidBz})
	if err != nil {
		return nil, err
	}

	return &data.MsgStoreDataResponse{}, nil
}