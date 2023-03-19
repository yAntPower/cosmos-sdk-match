package simapp

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	cliTx "github.com/cosmos/cosmos-sdk/client/tx"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/client/cli"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
	"golang.org/x/net/context"
	"strconv"
)

func (app *SimApp) Match() func(price float32) {
	return func(price float32) {
		if price <= 0.000000001 {
			return
		}
		buyTxs, sellTxs := app.GetMemPoolTxs()
		btxsCount := len(buyTxs)
		stxsCount := len(sellTxs)
		if btxsCount <= 0 || stxsCount <= 0 {
			return
		}
		if btxsCount > stxsCount {
			btxsCount = stxsCount
		}
		var (
			newBtxs = make([]sdk.Tx, 0)
			newStxs = make([]sdk.Tx, 0)
		)
		for i := 0; i < btxsCount; i++ {
			btx := buyTxs[i].GetMsgs()[0].(sdk.TxSellBuy)
			stx := sellTxs[i].GetMsgs()[0].(sdk.TxSellBuy)
			if btx.GetPrice() == price && price == stx.GetPrice() {
				tradeNum := float32(0.0)
				if btx.GetQuantity() > stx.GetQuantity() {
					lave := btx.GetQuantity() - stx.GetQuantity()
					btx.SetQuantity(lave)
					tradeNum = stx.GetQuantity()
					newBtxs = append(newBtxs, btx)
				} else if btx.GetQuantity() < stx.GetQuantity() {
					lave := stx.GetQuantity() - btx.GetQuantity()
					stx.SetQuantity(lave)
					tradeNum = btx.GetQuantity()
					newStxs = append(newStxs, stx)
				} else if btx.GetQuantity() == stx.GetQuantity() {
					tradeNum = btx.GetQuantity()
				}
				if tradeNum <= float32(0.000000001) {
					continue
				}
				// 生成原生的交易
				app.buildTx(stx.GetFrom(), btx.GetFrom(), tradeNum, price)
			} else {
				newBtxs = append(newBtxs, btx)
				newStxs = append(newStxs, stx)
			}
		}
		app.SetMemPoolTxs(newBtxs, newStxs)
	}
}

func (app *SimApp) buildTx(sellAddrStr, buyAddrStr string, tradeNum, price float32) {
	buyAddr, err := sdk.AccAddressFromBech32(buyAddrStr)
	if err != nil {
		app.Logger().Error("转换buyAddr失败，err=", err.Error())
		//continue
	}
	sellAddr, err := sdk.AccAddressFromBech32(sellAddrStr)
	if err != nil {
		app.Logger().Error("转换selladdr失败，err=", err.Error())
		//continue
	}
	coinStrBtc := strconv.FormatFloat(float64(tradeNum), 'f', 6, 32) + sdk.BTC
	coinsBtc, err := sdk.ParseCoinsNormalized(coinStrBtc)
	if err != nil {
		app.Logger().Error("转换coinsBtc失败，err=", err.Error())
		//continue
	}
	coinStrUSDT := strconv.FormatFloat(float64(tradeNum*price), 'f', 6, 32) + sdk.USDT
	coinsUSDT, err := sdk.ParseCoinsNormalized(coinStrUSDT)
	if err != nil {
		app.Logger().Error("转换coinsUSDT失败，err=", err.Error())
		//continue
	}
	msgBtc := bank.NewMsgSend(sellAddr, buyAddr, coinsBtc)
	msgUSDT := bank.NewMsgSend(buyAddr, sellAddr, coinsUSDT)

	extraArgs := []string{
		//fmt.Sprintf(" --%s=%s ", flags.FlagBroadcastMode, flags.BroadcastSync),
		//fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		//fmt.Sprintf("--%s=%s ", flags.FlagBroadcastMode, flags.BroadcastSync),
		//fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin("photon", sdk.NewInt(10))).String()),
		//fmt.Sprintf("--%s=%s ", flags.FlagFrom, sellAddrStr),
		fmt.Sprintf("--%s=%s ", flags.FlagKeyringBackend, "test"),
		fmt.Sprintf("--%s=%s ", flags.FlagChainID, "srspoa"),
	}

	ctx := svrcmd.CreateExecuteContext(context.Background())
	cmd := cli.NewSendTxCmd()
	cmd.SetContext(ctx)

	cmd.SetArgs(append([]string{sellAddrStr, buyAddrStr, coinStrBtc}, extraArgs...))
	//f := cmd.Flags()
	cmd.Flags().Set(flags.FlagChainID, "srspoa")
	cmd.Flags().Set(flags.FlagKeyringBackend, "test")
	//cmd.Flags().Set(flags.FlagFrom, sellAddrStr)
	clientCtx, err := client.GetClientTxContext(cmd)
	if err != nil {
		app.Logger().Error("转换clientCtx失败，err=", err.Error())
	}
	//clientCtx = clientCtx.WithFrom(sellAddrStr).WithFromAddress(sellAddr)
	clientCtx = clientCtx.WithTxConfig(app.txConfig)

	txf := cliTx.NewFactoryCLI(clientCtx, cmd.Flags())
	txf = txf.WithTxConfig(app.txConfig)
	tx, err := txf.BuildUnsignedTx(msgBtc)
	if err != nil {
		app.Logger().Error("转换BuildUnsignedTx失败，err=", err.Error())
	}
	tx.SetMemo("MATCH")
	//err = cliTx.Sign(txf, sellAddrStr, tx, true)
	//if err != nil {
	//	app.Logger().Error("转换Sign失败，err=", err.Error())
	//}

	txBytes, err := clientCtx.TxConfig.TxEncoder()(tx.GetTx())
	if err != nil {
		app.Logger().Error("转换TxEncoder失败，err=", err.Error())
	}
	ntx, err := app.GetTxDecoder()(txBytes)
	if err != nil {
		app.Logger().Error("转换GetTxDecoder失败，err=", err.Error())
	}
	app.InsertTxToNoopPool(ntx)

	tx, err = txf.BuildUnsignedTx(msgUSDT)
	if err != nil {
		app.Logger().Error("转换BuildUnsignedTx(USDT)失败，err=", err.Error())
	}
	tx.SetMemo("MATCH")
	//err = cliTx.Sign(txf, buyAddrStr, tx, true)
	//if err != nil {
	//	app.Logger().Error("转换Sign(USDT)失败，err=", err.Error())
	//}

	txBytes, err = clientCtx.TxConfig.TxEncoder()(tx.GetTx())
	if err != nil {
		app.Logger().Error("转换TxEncoder(USDT)失败，err=", err.Error())
	}
	ntx, err = app.GetTxDecoder()(txBytes)
	if err != nil {
		app.Logger().Error("转换GetTxDecoder(msgUSDT)失败，err=", err.Error())
	}
	app.InsertTxToNoopPool(ntx)
}
