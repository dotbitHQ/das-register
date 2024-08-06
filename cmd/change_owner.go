package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/dotbitHQ/das-lib/common"
	"github.com/dotbitHQ/das-lib/core"
	"github.com/dotbitHQ/das-lib/txbuilder"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/nervosnetwork/ckb-sdk-go/address"
	"github.com/nervosnetwork/ckb-sdk-go/indexer"
	"github.com/nervosnetwork/ckb-sdk-go/rpc"
	"github.com/nervosnetwork/ckb-sdk-go/types"
	"github.com/urfave/cli/v2"
	"math"
	"os"
)

func changeContractOwner(ctx *cli.Context) error {
	indexUrl := ctx.String("index_url")
	ckbClient, err := rpc.DialWithIndexer(indexUrl, indexUrl)
	if err != nil {
		return err
	}
	log.Info("ckb node ok")

	netType := ctx.Int("env")
	// das init
	env := core.InitEnvOpt(netType,
		common.DasContractNameAlwaysSuccess,
		common.DasContractNameConfigCellType,
		common.DasContractNameDispatchCellType,
		common.DasContractNameAccountCellType,
		common.DasContractNameBalanceCellType,
		common.DasContractNameApplyRegisterCellType,
		common.DasContractNamePreAccountCellType,
		common.DasContractNameProposalCellType,
		common.DasContractNameIncomeCellType,
		common.DasContractNameAccountSaleCellType,
		common.DasContractNameReverseRecordCellType,
		common.DASContractNameOfferCellType,
		common.DASContractNameSubAccountCellType,
		common.DASContractNameEip712LibCellType,
		common.DasContractNameReverseRecordRootCellType,
		common.DasKeyListCellType,
		common.DasContractNameDpCellType,
		common.DasContractNameDidCellType,
	)
	ops := []core.DasCoreOption{
		core.WithClient(ckbClient),
		core.WithDasContractArgs(env.ContractArgs),
		core.WithDasContractCodeHash(env.ContractCodeHash),
		core.WithDasNetType(netType),
		core.WithTHQCodeHash(env.THQCodeHash),
	}
	dasCore := core.NewDasCore(ctxServer, &wgServer, ops...)
	dasCore.InitDasContract(env.MapContract)
	if err := dasCore.InitDasConfigCell(); err != nil {
		return err
	}
	if err := dasCore.InitDasSoScript(); err != nil {
		return err
	}

	multiSignAddress, err := address.Parse(ctx.String("old_address"))
	if err != nil {
		return err
	}
	newMultiSignAddress, err := address.Parse(ctx.String("new_address"))
	if err != nil {
		return err
	}

	totalCells := 0
	totalNormalCells := make([]*indexer.LiveCell, 0)
	totalDataCells := 0
	totalUnkownCells := 0
	totalDidContractCells := 0
	totalDidNormalCells := 0
	normalCellsMap := make(map[string]int)

	searchKey := &indexer.SearchKey{
		Script:     multiSignAddress.Script,
		ScriptType: indexer.ScriptTypeLock,
	}
	nextCursor := ""

	for {
		cells, err := ckbClient.GetCells(context.Background(), searchKey, indexer.SearchOrderAsc, 10000, nextCursor)
		if err != nil {
			return err
		}
		if len(cells.Objects) == 0 {
			break
		}
		totalCells += len(cells.Objects)
		nextCursor = cells.LastCursor

		for _, cell := range cells.Objects {
			outpoint := common.OutPointStruct2String(cell.OutPoint)

			if cell.Output.Type == nil {
				if hex.EncodeToString(cell.OutputData) != "" {
					totalDataCells++
					data := cell.OutputData
					if len(cell.OutputData) > 66 {
						data = cell.OutputData[:66]
						fmt.Printf("%s: data cell, data: %s ...", outpoint, hex.EncodeToString(data))
					} else {
						fmt.Printf("%s: data cell, data: %s", outpoint, hex.EncodeToString(data))
					}
					continue
				}

				totalNormalCells = append(totalNormalCells, cell)
				continue
			}

			didCell := false

			// so script
			core.DasSoScriptMap.Range(func(key, value any) bool {
				item, ok := value.(*core.SoScript)
				if !ok {
					return true
				}
				if item.OutPut.Type.Equals(cell.Output.Type) {
					didCell = true
					totalDidContractCells++
					itemOutpoint := common.OutPointStruct2String(&item.OutPoint)
					if itemOutpoint == outpoint {
						fmt.Printf("%s: did contract cell (so script)[active], name: %s", outpoint, key)
					} else {
						fmt.Printf("%s: did contract cell (so script)[old], name: %s", outpoint, key)
					}
				}
				return true
			})

			// das contract cell
			core.DasContractMap.Range(func(key, value any) bool {
				item, ok := value.(*core.DasContractInfo)
				if !ok {
					return true
				}
				itemOutpoint := common.OutPointStruct2String(item.OutPoint)

				// cell type
				if item.OutPut.Type.Equals(cell.Output.Type) {
					didCell = true
					totalDidContractCells++
					if outpoint == itemOutpoint {
						fmt.Printf("%s: contract cell-type[active]: %s", outpoint, key)
					} else {
						fmt.Printf("%s: contract cell-type[old]: %s", outpoint, key)
					}
					return true
				}

				// cell type id
				if item.IsSameTypeId(cell.Output.Type.CodeHash) {
					didCell = true
					totalDidNormalCells++
					log.Infof("%s: cell name: %s", outpoint, key)

					contractName := string(item.ContractName)
					normalCellsMap[contractName]++
				}
				return true
			})

			if cell.Output.Type.CodeHash.String() == env.ContractCodeHash && !didCell {
				didCell = true
				totalDidNormalCells++
				contractName := "ContractSourceCell"

				fmt.Printf("%s: contract cell name: %s", outpoint, contractName)
				normalCellsMap[contractName]++
			}

			if !didCell {
				totalUnkownCells++
				fmt.Printf("%s: unkown cell, type: %s", outpoint, hex.EncodeToString(cell.Output.Type.CodeHash[:]))
			}
		}
	}

	fmt.Printf("total did contract cells: %d", totalDidContractCells)
	fmt.Printf("total did normal cells: %d, map: %s", len(totalNormalCells), gconv.String(normalCellsMap))
	fmt.Printf("total unkown cells: %d", totalUnkownCells)
	fmt.Printf("total data cells: %d", totalDataCells)
	fmt.Printf("total cells: %d", totalCells)

	// normalCellsTx
	stepCells := 2000
	for i := 0; len(totalNormalCells) > 0; i++ {
		var cells []*indexer.LiveCell
		if len(totalNormalCells) <= stepCells {
			cells = totalNormalCells
			totalNormalCells = nil
		} else {
			cells = totalNormalCells[:stepCells]
			totalNormalCells = totalNormalCells[stepCells:]
		}

		var txParams txbuilder.BuildTransactionParams
		soMulti, err := core.GetDasSoScript(common.SoScriptTypeCkbMulti)
		if err != nil {
			return err
		}
		txParams.CellDeps = append(txParams.CellDeps, soMulti.ToCellDep())
		for _, cell := range cells {
			txParams.Inputs = append(txParams.Inputs, &types.CellInput{
				PreviousOutput: cell.OutPoint,
			})
			txParams.Outputs = append(txParams.Outputs, &types.CellOutput{
				Capacity: cell.Output.Capacity,
				Lock:     newMultiSignAddress.Script,
			})
			txParams.OutputsData = append(txParams.OutputsData, []byte{})
		}

		base := txbuilder.NewDasTxBuilderBase(context.Background(), dasCore, nil, "")
		txBuilder := txbuilder.NewDasTxBuilderFromBase(base, nil)
		if err := txBuilder.BuildTransaction(&txParams); err != nil {
			return err
		}
		sizeInBlock, _ := txBuilder.Transaction.SizeInBlock()
		fee := uint64(math.Ceil(float64(sizeInBlock*1000) / float64(1000)))
		latestIndex := len(txBuilder.Transaction.Outputs) - 1
		changeCapacity := txBuilder.Transaction.Outputs[latestIndex].Capacity - fee
		txBuilder.Transaction.Outputs[latestIndex].Capacity = changeCapacity

		fileName := fmt.Sprintf("normalCellsTx_%d.json", i)
		if err := os.WriteFile(fileName, []byte(txBuilder.TxString()), 0666); err != nil {
			return err
		}
		fmt.Printf("write file: %s", fileName)
	}
	return nil
}
