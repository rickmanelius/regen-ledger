package server

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/regen-network/regen-ledger/orm"
	servermodule "github.com/regen-network/regen-ledger/types/module/server"
	"github.com/regen-network/regen-ledger/x/data"
	"github.com/regen-network/regen-ledger/x/ecocredit"
	"github.com/regen-network/regen-ledger/x/group"
	"github.com/regen-network/regen-ledger/x/group/exported"
)

const (
	// Group Table
	GroupTablePrefix        byte = 0x0
	GroupTableSeqPrefix     byte = 0x1
	GroupByAdminIndexPrefix byte = 0x2

	// Group Member Table
	GroupMemberTablePrefix         byte = 0x10
	GroupMemberByGroupIndexPrefix  byte = 0x11
	GroupMemberByMemberIndexPrefix byte = 0x12

	// Group Account Table
	GroupAccountTablePrefix        byte = 0x20
	GroupAccountTableSeqPrefix     byte = 0x21
	GroupAccountByGroupIndexPrefix byte = 0x22
	GroupAccountByAdminIndexPrefix byte = 0x23

	// Proposal Table
	ProposalTablePrefix               byte = 0x30
	ProposalTableSeqPrefix            byte = 0x31
	ProposalByGroupAccountIndexPrefix byte = 0x32
	ProposalByProposerIndexPrefix     byte = 0x33

	// Vote Table
	VoteTablePrefix           byte = 0x40
	VoteByProposalIndexPrefix byte = 0x41
	VoteByVoterIndexPrefix    byte = 0x42
)

type serverImpl struct {
	key servermodule.RootModuleKey

	accKeeper  exported.AccountKeeper
	bankKeeper exported.BankKeeper

	// Group Table
	groupTable        orm.AutoUInt64Table
	groupByAdminIndex orm.Index

	// Group Member Table
	groupMemberTable         orm.PrimaryKeyTable
	groupMemberByGroupIndex  orm.UInt64Index
	groupMemberByMemberIndex orm.Index

	// Group Account Table
	groupAccountSeq          orm.Sequence
	groupAccountTable        orm.PrimaryKeyTable
	groupAccountByGroupIndex orm.UInt64Index
	groupAccountByAdminIndex orm.Index

	// Proposal Table
	proposalTable               orm.AutoUInt64Table
	proposalByGroupAccountIndex orm.Index
	proposalByProposerIndex     orm.Index

	// Vote Table
	voteTable           orm.PrimaryKeyTable
	voteByProposalIndex orm.UInt64Index
	voteByVoterIndex    orm.Index
}

func newServer(storeKey servermodule.RootModuleKey, accKeeper exported.AccountKeeper, bankKeeper exported.BankKeeper, cdc codec.Codec) serverImpl {
	s := serverImpl{key: storeKey, accKeeper: accKeeper, bankKeeper: bankKeeper}

	// Group Table
	groupTableBuilder, err := orm.NewAutoUInt64TableBuilder(GroupTablePrefix, GroupTableSeqPrefix, storeKey, &group.GroupInfo{}, cdc)
	if err != nil {
		panic(err.Error())
	}
	s.groupByAdminIndex, err = orm.NewIndex(groupTableBuilder, GroupByAdminIndexPrefix, func(val interface{}) ([]orm.RowID, error) {
		addr, err := sdk.AccAddressFromBech32(val.(*group.GroupInfo).Admin)
		if err != nil {
			return nil, err
		}
		return []orm.RowID{addr.Bytes()}, nil
	})
	if err != nil {
		panic(err.Error())
	}
	s.groupTable = groupTableBuilder.Build()

	// Group Member Table
	groupMemberTableBuilder, err := orm.NewPrimaryKeyTableBuilder(GroupMemberTablePrefix, storeKey, &group.GroupMember{}, orm.Max255DynamicLengthIndexKeyCodec{}, cdc)
	if err != nil {
		panic(err.Error())
	}
	s.groupMemberByGroupIndex, err = orm.NewUInt64Index(groupMemberTableBuilder, GroupMemberByGroupIndexPrefix, func(val interface{}) ([]uint64, error) {
		group := val.(*group.GroupMember).GroupId
		return []uint64{group}, nil
	})
	if err != nil {
		panic(err.Error())
	}
	s.groupMemberByMemberIndex, err = orm.NewIndex(groupMemberTableBuilder, GroupMemberByMemberIndexPrefix, func(val interface{}) ([]orm.RowID, error) {
		memberAddr := val.(*group.GroupMember).Member.Address
		addr, err := sdk.AccAddressFromBech32(memberAddr)
		if err != nil {
			return nil, err
		}
		return []orm.RowID{addr.Bytes()}, nil
	})
	if err != nil {
		panic(err.Error())
	}
	s.groupMemberTable = groupMemberTableBuilder.Build()

	// Group Account Table
	s.groupAccountSeq = orm.NewSequence(storeKey, GroupAccountTableSeqPrefix)
	groupAccountTableBuilder, err := orm.NewPrimaryKeyTableBuilder(GroupAccountTablePrefix, storeKey, &group.GroupAccountInfo{}, orm.Max255DynamicLengthIndexKeyCodec{}, cdc)
	if err != nil {
		panic(err.Error())
	}
	s.groupAccountByGroupIndex, err = orm.NewUInt64Index(groupAccountTableBuilder, GroupAccountByGroupIndexPrefix, func(value interface{}) ([]uint64, error) {
		group := value.(*group.GroupAccountInfo).GroupId
		return []uint64{group}, nil
	})
	if err != nil {
		panic(err.Error())
	}
	s.groupAccountByAdminIndex, err = orm.NewIndex(groupAccountTableBuilder, GroupAccountByAdminIndexPrefix, func(value interface{}) ([]orm.RowID, error) {
		admin := value.(*group.GroupAccountInfo).Admin
		addr, err := sdk.AccAddressFromBech32(admin)
		if err != nil {
			return nil, err
		}
		return []orm.RowID{addr.Bytes()}, nil
	})
	if err != nil {
		panic(err.Error())
	}
	s.groupAccountTable = groupAccountTableBuilder.Build()

	// Proposal Table
	proposalTableBuilder, err := orm.NewAutoUInt64TableBuilder(ProposalTablePrefix, ProposalTableSeqPrefix, storeKey, &group.Proposal{}, cdc)
	if err != nil {
		panic(err.Error())
	}
	s.proposalByGroupAccountIndex, err = orm.NewIndex(proposalTableBuilder, ProposalByGroupAccountIndexPrefix, func(value interface{}) ([]orm.RowID, error) {
		account := value.(*group.Proposal).Address
		addr, err := sdk.AccAddressFromBech32(account)
		if err != nil {
			return nil, err
		}
		return []orm.RowID{addr.Bytes()}, nil
	})
	if err != nil {
		panic(err.Error())
	}
	s.proposalByProposerIndex, err = orm.NewIndex(proposalTableBuilder, ProposalByProposerIndexPrefix, func(value interface{}) ([]orm.RowID, error) {
		proposers := value.(*group.Proposal).Proposers
		r := make([]orm.RowID, len(proposers))
		for i := range proposers {
			addr, err := sdk.AccAddressFromBech32(proposers[i])
			if err != nil {
				return nil, err
			}
			r[i] = addr.Bytes()
		}
		return r, nil
	})
	if err != nil {
		panic(err.Error())
	}
	s.proposalTable = proposalTableBuilder.Build()

	// Vote Table
	voteTableBuilder, err := orm.NewPrimaryKeyTableBuilder(VoteTablePrefix, storeKey, &group.Vote{}, orm.Max255DynamicLengthIndexKeyCodec{}, cdc)
	if err != nil {
		panic(err.Error())
	}
	s.voteByProposalIndex, err = orm.NewUInt64Index(voteTableBuilder, VoteByProposalIndexPrefix, func(value interface{}) ([]uint64, error) {
		return []uint64{value.(*group.Vote).ProposalId}, nil
	})
	if err != nil {
		panic(err.Error())
	}
	s.voteByVoterIndex, err = orm.NewIndex(voteTableBuilder, VoteByVoterIndexPrefix, func(value interface{}) ([]orm.RowID, error) {
		addr, err := sdk.AccAddressFromBech32(value.(*group.Vote).Voter)
		if err != nil {
			return nil, err
		}
		return []orm.RowID{addr.Bytes()}, nil
	})
	if err != nil {
		panic(err.Error())
	}
	s.voteTable = voteTableBuilder.Build()

	return s
}

func RegisterServices(configurator servermodule.Configurator, accountKeeper exported.AccountKeeper, bankKeeper exported.BankKeeper) {
	impl := newServer(configurator.ModuleKey(), accountKeeper, bankKeeper, configurator.Marshaler())
	group.RegisterMsgServer(configurator.MsgServer(), impl)
	group.RegisterQueryServer(configurator.QueryServer(), impl)
	configurator.RegisterInvariantsHandler(impl.RegisterInvariants)
	configurator.RegisterGenesisHandlers(impl.InitGenesis, impl.ExportGenesis)
	configurator.RegisterWeightedOperationsHandler(impl.WeightedOperations)

	// Require servers from external modules for ADR 033 message routing
	configurator.RequireServer((*ecocredit.MsgServer)(nil))
	configurator.RequireServer((*data.MsgServer)(nil))
}
