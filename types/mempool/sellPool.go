package mempool

import (
	"context"
	"errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ Mempool = (*SellMempool)(nil)

type SellMempool struct {
	Txs []sdk.Tx
}
type SellMempoolOption func(*SellMempool)

func NewSellMempool() Mempool {
	mp := &SellMempool{
		Txs: make([]sdk.Tx, 0),
	}

	return mp
}

func (mp *SellMempool) Insert(ctx context.Context, tx sdk.Tx) error {
	mp.Txs = append(mp.Txs, tx)
	mp.Txs = TxSort(true, mp.Txs)
	return nil
}
func (mp *SellMempool) Select(context.Context, [][]byte) Iterator { return nil }
func (mp *SellMempool) CountTx() int                              { return len(mp.Txs) }
func (mp *SellMempool) Remove(sendTx sdk.Tx) error {
	for i, mt := range mp.Txs {
		poolTx := mt.GetMsgs()[0].(sdk.TxSellBuy)
		tx := sendTx.GetMsgs()[0].(sdk.TxSellBuy)
		if poolTx.GetPrice() == tx.GetPrice() && poolTx.GetQuantity() == tx.GetQuantity() {
			if i < len(mp.Txs)-1 {
				mp.Txs = append(mp.Txs[:i], mp.Txs[i+1:]...)
			} else if i == len(mp.Txs)-1 {
				mp.Txs = mp.Txs[:i]
			} else {
				err := errors.New("emit macho dwarf: elf header corrupted")
				return err
			}
			break
		}
	}
	return nil
}

func (mp *SellMempool) GetTxsByPool() []sdk.Tx {
	return mp.Txs
}
func (mp *SellMempool) SetTxsToPool(txs []sdk.Tx) {
	mp.Txs = TxSort(true, txs)
}
