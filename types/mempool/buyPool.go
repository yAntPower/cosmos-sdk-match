package mempool

import (
	"context"
	"errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ Mempool = (*BuyMempool)(nil)

type BuyMempool struct {
	Txs []sdk.Tx
}

func NewBuyMempool() Mempool {
	mp := &BuyMempool{
		Txs: make([]sdk.Tx, 0),
	}

	return mp
}

func (mp *BuyMempool) Insert(ctx context.Context, tx sdk.Tx) error {
	mp.Txs = append(mp.Txs, tx)
	mp.Txs = TxSort(false, mp.Txs)
	return nil
}
func (mp *BuyMempool) Select(context.Context, [][]byte) Iterator {
	return nil
}
func (mp *BuyMempool) CountTx() int {
	return len(mp.Txs)
}
func (mp *BuyMempool) Remove(sendTx sdk.Tx) error {
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
func (mp *BuyMempool) GetTxsByPool() []sdk.Tx {
	return mp.Txs
}
func (mp *BuyMempool) SetTxsToPool(txs []sdk.Tx) {
	mp.Txs = TxSort(false, txs)
}
