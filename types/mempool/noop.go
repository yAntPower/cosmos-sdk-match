package mempool

import (
	"context"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ Mempool = (*NoOpMempool)(nil)

// NoOpMempool defines a no-op mempool. Transactions are completely discarded and
// ignored when BaseApp interacts with the mempool.
//
// Note: When this mempool is used, it assumed that an application will rely
// on Tendermint's transaction ordering defined in `RequestPrepareProposal`, which
// is FIFO-ordered by default.
type NoOpMempool struct {
	Txs []sdk.Tx
}

//type NoOpMempoolOption func(*NoOpMempool)

func NewNoOpMempool() Mempool {
	mp := &NoOpMempool{
		Txs: make([]sdk.Tx, 0),
	}

	return mp
}

func (mp *NoOpMempool) Insert(ctx context.Context, tx sdk.Tx) error {
	mp.Txs = append(mp.Txs, tx)
	return nil
}

func (mp *NoOpMempool) Select(context.Context, [][]byte) Iterator { return nil }
func (mp *NoOpMempool) CountTx() int                              { return 0 }
func (mp *NoOpMempool) Remove(sendTx sdk.Tx) error {
	if len(mp.Txs) > 1 {
		mp.Txs = mp.Txs[1:]
	} else {
		mp.Txs = make([]sdk.Tx, 0)
	}
	//for i, mt := range mp.Txs {
	//	poolTx := mt.GetMsgs()[0].(sdk.TxSellBuy)
	//	tx := sendTx.GetMsgs()[0].(sdk.TxSellBuy)
	//	if poolTx.GetPrice() == tx.GetPrice() && poolTx.GetQuantity() == tx.GetQuantity() {
	//		if i < len(mp.Txs)-1 {
	//			mp.Txs = append(mp.Txs[:i], mp.Txs[i+1:]...)
	//		} else if i == len(mp.Txs)-1 {
	//			mp.Txs = mp.Txs[:i]
	//		} else {
	//			err := errors.New("emit macho dwarf: elf header corrupted")
	//			return err
	//		}
	//		break
	//	}
	//}
	return nil
}
func (mp *NoOpMempool) GetTxsByPool() []sdk.Tx {
	return mp.Txs
}
func (mp *NoOpMempool) SetTxsToPool(txs []sdk.Tx) {
	//mp.Txs = txs
}
