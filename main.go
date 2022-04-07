package main

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"github.com/laizy/web3"
	ontology_go_sdk "github.com/ontio/ontology-go-sdk"
	"github.com/ontio/ontology-go-sdk/oep4"
	"github.com/ontio/ontology/core/types"
	"math/big"
	"path"
	"strings"

	//common3 "github.com/ontio/ontology-go-sdk/common"
	common2 "github.com/ontio/ontology/common"
	"github.com/ontio/ontology/smartcontract/service/native/ont"
	"github.com/transerTools/common"
	"github.com/transerTools/config"
	"github.com/transerTools/utils"
	"github.com/xuri/excelize/v2"
	"os"
	"strconv"
	"time"
)

func main() {
	cf := config.ParseConfig()

	sdk := ontology_go_sdk.NewOntologySdk()
	sdk.NewRpcClient().SetAddress(cf.RpcUrl)

	wall, err := sdk.OpenWallet(cf.WalletFile)
	common.CheckErr(err)
	admin, err := wall.GetAccountByAddress(cf.Address, []byte(cf.Password))
	common.CheckErr(err)
	fmt.Println("from address:", admin.Address.ToBase58())

	contractAddr := cf.GetContractAddress()
	var decimals = getDecimals(contractAddr, sdk)

	toInfos, sum, fee := parseExcel(cf, decimals)
	checkBalance(contractAddr, sdk, admin.Address, sum, fee)
	if cf.Execute {
		transfer(contractAddr, sdk, admin, toInfos, cf)
	}
}

func transfer(contractAddr common2.Address, sdk *ontology_go_sdk.OntologySdk, admin *ontology_go_sdk.Account, toInfos []*config.ToInfo, cf *config.Config) {
	now := time.Now().Unix()

	var wBackup *bufio.Writer
	fileName := path.Base(cf.ExcelFile)
	fn := strings.Split(fileName, ".")
	if len(fn) < 2 {
		panic(fileName)
	}
	if cf.Execute {
		f2, err := os.OpenFile(fn[0]+"_"+strconv.Itoa(int(now))+".txt", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
		common.CheckErr(err)
		wBackup = bufio.NewWriter(f2)
		defer func() {
			wBackup.Flush()
			f2.Sync()
			f2.Close()
		}()
	}
	isONG := contractAddr == ontology_go_sdk.ONG_CONTRACT_ADDRESS
	isONT := contractAddr == ontology_go_sdk.ONT_CONTRACT_ADDRESS
	if isONG || isONT {
		states := make([]*ont.State, 0)
		end := len(toInfos)
		for k, toInfo := range toInfos {
			if len(states) >= 20 || k == end-1 {
				var err error
				var tx *types.MutableTransaction
				if isONG {
					tx, err = sdk.Native.Ong.NewMultiTransferTransaction(cf.GasPrice, cf.GasLimit, states)
				} else if isONT {
					tx, err = sdk.Native.Ont.NewMultiTransferTransaction(cf.GasPrice, cf.GasLimit, states)
				} else {
					panic("invalid type")
				}
				common.CheckErr(err)
				tx.Payer = admin.Address
				err = sdk.SignToTransaction(tx, admin)
				common.CheckErr(err)
				imutTx, err := tx.IntoImmutable()
				common.CheckErr(err)
				txHash := imutTx.Hash()
				wBackup.WriteString(fmt.Sprintf("txHash:%s, txHex:%s", txHash.ToHexString(), hex.EncodeToString(imutTx.ToArray())))
				wBackup.Write([]byte("\n"))
				_, err = sdk.SendTransaction(tx)
				common.CheckErr(err)
				states = make([]*ont.State, 0)
			} else {
				state := &ont.State{
					From:  admin.Address,
					To:    toInfo.To,
					Value: toInfo.Amount.Uint64(),
				}
				states = append(states, state)
			}
		}
	} else {
		oep4States := make([][]interface{}, 0)
		end := len(toInfos)
		for k, toInfo := range toInfos {
			if len(oep4States) >= 20 || k == end-1 {
				var txHash common2.Uint256
				var err error
				mutableTx, err := sdk.NeoVM.NewNeoVMInvokeTransaction(cf.GasPrice, cf.GasLimit, contractAddr,
					[]interface{}{"transferMulti", oep4States})
				common.CheckErr(err)
				mutableTx.Payer = admin.Address
				err = sdk.SignToTransaction(mutableTx, admin)
				common.CheckErr(err)
				imutTx, err := mutableTx.IntoImmutable()
				common.CheckErr(err)
				txHash = imutTx.Hash()
				wBackup.WriteString(fmt.Sprintf("txHash:%s, txHex:%s", txHash.ToHexString(), hex.EncodeToString(imutTx.ToArray())))
				wBackup.Write([]byte("\n"))

				txHash, err = sdk.SendTransaction(mutableTx)
				common.CheckErr(err)
				oep4States = make([][]interface{}, 0)
			} else {
				state := []interface{}{admin.Address, toInfo.To, toInfo.Amount}
				oep4States = append(oep4States, state)
			}
		}
	}
}

func checkBalance(contractAddr common2.Address, sdk *ontology_go_sdk.OntologySdk, admin common2.Address,
	sum *big.Int, fee uint64) {

	if contractAddr == ontology_go_sdk.ONG_CONTRACT_ADDRESS {
		bal, err := sdk.Native.Ong.BalanceOf(admin)
		common.CheckErr(err)
		if bal < sum.Uint64()+fee {
			panic("ONG: Insufficient balance")
		}
	} else if contractAddr == ontology_go_sdk.ONT_CONTRACT_ADDRESS {
		bal, err := sdk.Native.Ont.BalanceOf(admin)
		common.CheckErr(err)
		if bal < sum.Uint64() {
			panic("ONT: Insufficient balance")
		}
		bal, err = sdk.Native.Ong.BalanceOf(admin)
		common.CheckErr(err)
		if bal < fee {
			panic("ONG: Insufficient balance")
		}
	} else {
		oep4Token := oep4.NewOep4(contractAddr, sdk)
		bal, err := oep4Token.BalanceOf(admin)
		common.CheckErr(err)
		if bal.Cmp(sum) < 0 {
			panic("OEP4: Insufficient balance")
		}
		ongBal, err := sdk.Native.Ong.BalanceOf(admin)
		common.CheckErr(err)
		if ongBal < fee {
			panic("ONG: Insufficient balance")
		}
	}
}

func getDecimals(contractAddr common2.Address, sdk *ontology_go_sdk.OntologySdk) uint64 {
	var decimals uint64
	if contractAddr == ontology_go_sdk.ONG_CONTRACT_ADDRESS {
		decimals = 9
	} else if contractAddr == ontology_go_sdk.ONT_CONTRACT_ADDRESS {
		decimals = 0
	} else {
		oep4Token := oep4.NewOep4(contractAddr, sdk)
		dec, err := oep4Token.Decimals()
		common.CheckErr(err)
		decimals = dec.Uint64()
	}
	return decimals
}

func parseExcel(cf *config.Config, decimals uint64) ([]*config.ToInfo, *big.Int, uint64) {
	f, err := excelize.OpenFile(cf.ExcelFile)
	common.CheckErr(err)
	// 获取 Sheet1 上所有单元格
	rows, err := f.GetRows("Sheet1")
	common.CheckErr(err)
	var toInfos = make([]*config.ToInfo, 0)
	var isBase58 bool
	for i := 1; i < len(rows); i++ {
		if rows[i][0] == "" {
			break
		}
		var to common2.Address
		to, err := common2.AddressFromBase58(rows[i][0])
		if err != nil {
			if !strings.HasPrefix(rows[i][0], "0x") {
				rows[i][0] = "0x" + rows[i][0]
			}
			ethAdr := web3.HexToAddress(rows[i][0])
			to = common2.Address(ethAdr)
			isBase58 = false
		} else {
			isBase58 = true
		}
		toInfos = append(toInfos, &config.ToInfo{
			To:     to,
			Amount: utils.ToIntByPrecise(rows[i][1], decimals),
		})
	}
	sum := big.NewInt(0)

	for _, toInfo := range toInfos {
		sum = big.NewInt(0).Add(sum, toInfo.Amount)
		if !cf.Execute {
			if isBase58 {
				fmt.Println("to:", toInfo.To.ToBase58(), "amount:", toInfo.Amount.String())
			} else {
				fmt.Println("to:", "0x"+hex.EncodeToString(toInfo.To[:]), "amount:", toInfo.Amount.String())
			}
		}
	}

	var fee uint64
	if cf.GetContractAddress() == ontology_go_sdk.ONG_CONTRACT_ADDRESS || cf.GetContractAddress() == ontology_go_sdk.ONT_CONTRACT_ADDRESS {
		fee = uint64(len(toInfos)/20+1) * 120000000
	} else {
		fee = uint64(len(toInfos)/20+1) * 550000000
	}
	if !cf.Execute {
		fmt.Println("estimate tx fee is: ", float64(fee)/1000000000)
	}
	return toInfos, sum, fee
}
