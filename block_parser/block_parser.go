package block_parser

import (
	"context"
	"das_register_server/config"
	"das_register_server/dao"
	"das_register_server/notify"
	"das_register_server/tables"
	"fmt"
	"github.com/DeAccountSystems/das-lib/common"
	"github.com/DeAccountSystems/das-lib/core"
	"github.com/DeAccountSystems/das-lib/witness"
	"github.com/nervosnetwork/ckb-sdk-go/types"
	"github.com/scorpiotzh/mylog"
	"sync"
	"sync/atomic"
	"time"
)

var log = mylog.NewLogger("block_parser", mylog.LevelDebug)

type BlockParser struct {
	DasCore              *core.DasCore
	mapTransactionHandle map[common.DasAction]FuncTransactionHandle
	CurrentBlockNumber   uint64
	DbDao                *dao.DbDao
	ConcurrencyNum       uint64
	ConfirmNum           uint64
	Ctx                  context.Context
	Wg                   *sync.WaitGroup
}

func (b *BlockParser) Run() error {
	b.registerTransactionHandle()
	currentBlockNumber, err := b.DasCore.Client().GetTipBlockNumber(b.Ctx)
	if err != nil {
		return fmt.Errorf("GetTipBlockNumber err: %s", err.Error())
	}

	if err := b.initCurrentBlockNumber(currentBlockNumber); err != nil {
		return fmt.Errorf("initCurrentBlockNumber err: %s", err.Error())
	}
	atomic.AddUint64(&b.CurrentBlockNumber, 1)
	b.Wg.Add(1)
	go func() {
		for {
			select {
			default:
				latestBlockNumber, err := b.DasCore.Client().GetTipBlockNumber(b.Ctx)
				if err != nil {
					log.Error("GetTipBlockNumber err:", err.Error())
				} else {
					if b.ConcurrencyNum > 1 && b.CurrentBlockNumber < (latestBlockNumber-b.ConfirmNum-b.ConcurrencyNum) {
						nowTime := time.Now()
						if err = b.parserConcurrencyMode(); err != nil {
							log.Error("parserConcurrencyMode err:", err.Error(), b.CurrentBlockNumber)
						}
						log.Warn("parserConcurrencyMode time:", time.Since(nowTime).Seconds())
					} else if b.CurrentBlockNumber < (latestBlockNumber - b.ConfirmNum) { // check rollback
						nowTime := time.Now()
						if err = b.parserSubMode(); err != nil {
							log.Error("parserSubMode err:", err.Error(), b.CurrentBlockNumber)
						}
						log.Warn("parserSubMode time:", time.Since(nowTime).Seconds())
					} else {
						log.Info("RunParser:", b.CurrentBlockNumber, latestBlockNumber)
						time.Sleep(time.Second * 10)
					}
					time.Sleep(time.Millisecond * 300)
				}
			case <-b.Ctx.Done():
				b.Wg.Done()
				return
			}
		}
	}()
	return nil
}

func (b *BlockParser) initCurrentBlockNumber(currentBlockNumber uint64) error {
	if block, err := b.DbDao.FindBlockInfo(tables.ParserTypeDAS); err != nil {
		return err
	} else if block.Id > 0 {
		b.CurrentBlockNumber = block.BlockNumber
	} else if b.CurrentBlockNumber == 0 && currentBlockNumber > 0 {
		b.CurrentBlockNumber = currentBlockNumber
	}
	return nil
}

func (b *BlockParser) parserSubMode() error {
	log.Info("parserSubMode:", b.CurrentBlockNumber)
	block, err := b.DasCore.Client().GetBlockByNumber(b.Ctx, b.CurrentBlockNumber)
	if err != nil {
		return fmt.Errorf("GetBlockByNumber err: %s", err.Error())
	} else {
		blockHash := block.Header.Hash.Hex()
		parentHash := block.Header.ParentHash.Hex()
		log.Info("parserSubMode:", b.CurrentBlockNumber, blockHash, parentHash)
		if fork, err := b.checkFork(parentHash); err != nil {
			return fmt.Errorf("checkFork err: %s", err.Error())
		} else if fork {
			log.Warn("CheckFork is true:", b.CurrentBlockNumber, blockHash, parentHash)
			atomic.AddUint64(&b.CurrentBlockNumber, ^uint64(0))
		} else if err = b.parsingBlockData(block); err != nil {
			return fmt.Errorf("parsingBlockData err: %s", err.Error())
		} else {
			if err = b.DbDao.CreateBlockInfo(tables.ParserTypeDAS, b.CurrentBlockNumber, blockHash, parentHash); err != nil {
				return fmt.Errorf("CreateBlockInfo err: %s", err.Error())
			} else {
				atomic.AddUint64(&b.CurrentBlockNumber, 1)
			}
			if err = b.DbDao.DeleteBlockInfo(tables.ParserTypeDAS, b.CurrentBlockNumber-20); err != nil {
				return fmt.Errorf("DeleteBlockInfo err: %s", err.Error())
			}
		}
	}
	return nil
}

func (b *BlockParser) checkFork(parentHash string) (bool, error) {
	block, err := b.DbDao.FindBlockInfoByBlockNumber(tables.ParserTypeDAS, b.CurrentBlockNumber-1)
	if err != nil {
		return false, err
	}
	if block.Id == 0 {
		return false, nil
	} else if block.BlockHash != parentHash {
		log.Warn("CheckFork:", b.CurrentBlockNumber, parentHash, block.BlockHash)
		return true, nil
	}
	return false, nil
}

func (b *BlockParser) parserConcurrencyMode() error {
	log.Info("parserConcurrencyMode:", b.CurrentBlockNumber, b.ConcurrencyNum)
	for i := uint64(0); i < b.ConcurrencyNum; i++ {
		block, err := b.DasCore.Client().GetBlockByNumber(b.Ctx, b.CurrentBlockNumber)
		if err != nil {
			return fmt.Errorf("GetBlockByNumber err: %s [%d]", err.Error(), b.CurrentBlockNumber)
		}
		blockHash := block.Header.Hash.Hex()
		parentHash := block.Header.ParentHash.Hex()
		log.Info("parserConcurrencyMode:", b.CurrentBlockNumber, blockHash, parentHash)

		if err = b.parsingBlockData(block); err != nil {
			return fmt.Errorf("parsingBlockData err: %s", err.Error())
		} else {
			if err = b.DbDao.CreateBlockInfo(tables.ParserTypeDAS, b.CurrentBlockNumber, blockHash, parentHash); err != nil {
				return fmt.Errorf("CreateBlockInfo err: %s", err.Error())
			} else {
				atomic.AddUint64(&b.CurrentBlockNumber, 1)
			}
		}
	}
	if err := b.DbDao.DeleteBlockInfo(tables.ParserTypeDAS, b.CurrentBlockNumber-20); err != nil {
		return fmt.Errorf("DeleteBlockInfo err: %s", err.Error())
	}
	return nil
}

func (b *BlockParser) parsingBlockData(block *types.Block) error {
	for _, tx := range block.Transactions {
		txHash := tx.Hash.Hex()
		blockNumber := block.Header.Number
		blockTimestamp := block.Header.Timestamp

		if builder, err := witness.ActionDataBuilderFromTx(tx); err != nil {
			log.Warn("ActionDataBuilderFromTx err:", err.Error())
		} else {
			if handle, ok := b.mapTransactionHandle[builder.Action]; ok {
				// transaction parse by action
				resp := handle(FuncTransactionHandleReq{
					DbDao:          b.DbDao,
					Tx:             tx,
					TxHash:         txHash,
					BlockNumber:    blockNumber,
					BlockTimestamp: blockTimestamp,
					Action:         builder.Action,
				})
				if resp.Err != nil {
					log.Error("action handle resp:", builder.Action, blockNumber, txHash, resp.Err.Error())
					notify.SendLarkTextNotify(config.Cfg.Notify.LarkErrorKey, "Block Parse", notify.GetLarkTextNotifyStr("TransactionHandle", txHash, resp.Err.Error()))
					return resp.Err
				}
			}

		}
	}
	return nil
}
