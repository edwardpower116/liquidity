package liquidity_test

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/liquidity/app"
	"github.com/tendermint/liquidity/x/liquidity"
	"github.com/tendermint/liquidity/x/liquidity/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"testing"
)

func TestGenesisState(t *testing.T) {
	cdc := codec.NewLegacyAmino()
	types.RegisterLegacyAminoCodec(cdc)
	simapp := app.Setup(false)

	ctx := simapp.BaseApp.NewContext(false, tmproto.Header{})
	genesis := types.DefaultGenesisState()

	liquidity.InitGenesis(ctx, simapp.LiquidityKeeper, *genesis)

	defaultGenesisExported := liquidity.ExportGenesis(ctx, simapp.LiquidityKeeper)

	require.Equal(t, genesis, defaultGenesisExported)

	// define test denom X, Y for Liquidity Pool
	denomX, denomY := types.AlphabeticalDenomPair("denomX", "denomY")

	X := sdk.NewInt(1000000000)
	Y := sdk.NewInt(1000000000)

	addrs := app.AddTestAddrsIncremental(simapp, ctx, 20, sdk.NewInt(10000))
	poolId := app.TestCreatePool(t, simapp, ctx, X, Y, denomX, denomY, addrs[0])

	// begin block, init
	app.TestDepositPool(t, simapp, ctx, X.QuoRaw(10), Y, addrs[1:2], poolId, true)
	app.TestDepositPool(t, simapp, ctx, X, Y.QuoRaw(10), addrs[2:3], poolId, true)

	// next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	liquidity.BeginBlocker(ctx, simapp.LiquidityKeeper)

	price, _ := sdk.NewDecFromStr("1.1")
	offerCoins := []sdk.Coin{sdk.NewCoin(denomX, sdk.NewInt(10000))}
	orderPrices := []sdk.Dec{price}
	orderAddrs := addrs[1:2]
	_, _ = app.TestSwapPool(t, simapp, ctx, offerCoins, orderPrices, orderAddrs, poolId, false)
	_, _ = app.TestSwapPool(t, simapp, ctx, offerCoins, orderPrices, orderAddrs, poolId, false)
	_, _ = app.TestSwapPool(t, simapp, ctx, offerCoins, orderPrices, orderAddrs, poolId, false)
	_, _ = app.TestSwapPool(t, simapp, ctx, offerCoins, orderPrices, orderAddrs, poolId, true)

	genesisExported := liquidity.ExportGenesis(ctx, simapp.LiquidityKeeper)
	bankGenesisExported := simapp.BankKeeper.ExportGenesis(ctx)

	simapp2 := app.Setup(false)

	ctx2 := simapp2.BaseApp.NewContext(false, tmproto.Header{})
	ctx2 = ctx2.WithBlockHeight(1)

	simapp2.BankKeeper.InitGenesis(ctx2, bankGenesisExported)
	liquidity.InitGenesis(ctx2, simapp2.LiquidityKeeper, *genesisExported)
	simapp2GenesisExported := liquidity.ExportGenesis(ctx2, simapp2.LiquidityKeeper)
	require.Equal(t, genesisExported, simapp2GenesisExported)
}
