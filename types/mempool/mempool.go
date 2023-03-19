package mempool

import (
	"context"
	"errors"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"sort"
)

type Mempool interface {
	// Insert attempts to insert a Tx into the app-side mempool returning
	// an error upon failure.
	Insert(context.Context, sdk.Tx) error

	// Select returns an Iterator over the app-side mempool. If txs are specified,
	// then they shall be incorporated into the Iterator. The Iterator must
	// closed by the caller.
	Select(context.Context, [][]byte) Iterator

	// CountTx returns the number of transactions currently in the mempool.
	CountTx() int

	// Remove attempts to remove a transaction from the mempool, returning an error
	// upon failure.
	Remove(sdk.Tx) error

	GetTxsByPool() []sdk.Tx
	SetTxsToPool(txs []sdk.Tx)
}

// Iterator defines an app-side mempool iterator interface that is as minimal as possible.  The order of iteration
// is determined by the app-side mempool implementation.
type Iterator interface {
	// Next returns the next transaction from the mempool. If there are no more transactions, it returns nil.
	Next() Iterator

	// Tx returns the transaction at the current position of the iterator.
	Tx() sdk.Tx
}

var (
	ErrTxNotFound           = errors.New("tx not found in mempool")
	ErrMempoolTxMaxCapacity = errors.New("pool reached max tx capacity")
)

// Positive 是否是正序.
func TxSort(isPositive bool, txs []sdk.Tx) []sdk.Tx {
	sort.SliceStable(txs, func(i, j int) bool {
		msgsi := txs[i].GetMsgs()
		msgsj := txs[j].GetMsgs()
		if len(msgsi) <= 0 || len(msgsj) <= 0 {
			err := fmt.Sprintf("error：交易排序错误。是否是正序:%t", isPositive)
			panic(err)
		}
		if isPositive {
			return msgsi[0].(sdk.TxSellBuy).GetPrice() < msgsj[0].(sdk.TxSellBuy).GetPrice()
		} else {
			return msgsi[0].(sdk.TxSellBuy).GetPrice() > msgsj[0].(sdk.TxSellBuy).GetPrice()
		}
	})
	return txs
}
