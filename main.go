package main

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"github.com/laizy/web3"
	ontology_go_sdk "github.com/ontio/ontology-go-sdk"
	common3 "github.com/ontio/ontology-go-sdk/common"
	common2 "github.com/ontio/ontology/common"
	"github.com/transerTools/common"
	"github.com/transerTools/config"
	"github.com/transerTools/utils"
	"github.com/xuri/excelize/v2"
	"os"
	"strconv"
	"strings"
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

	if cf.GetContractAddress() != ontology_go_sdk.ONG_CONTRACT_ADDRESS {
		panic("only support ong")
	}
	toInfos, sum := parseExcel(cf)
	bal, err := sdk.Native.Ong.BalanceOf(admin.Address)
	common.CheckErr(err)
	if bal < sum {
		panic("Insufficient balance")
	}
	now := time.Now().Unix()

	var w *bufio.Writer
	if cf.Execute {
		f, err := os.OpenFile(strconv.Itoa(int(now))+".txt", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
		common.CheckErr(err)
		w = bufio.NewWriter(f)
		defer func() {
			w.Flush()
			f.Sync()
			f.Close()
		}()
	}
	if cf.Execute {
		states := make([]*common3.TransferState, 0)
		end := len(toInfos)
		for k, toInfo := range toInfos {
			if len(states) >= 20 || k == end-1 {
				txHash, err := sdk.Native.Ong.MultiTransfer(cf.GasPrice, cf.GasLimit, states, admin)
				if cf.Execute {
					w.WriteString(txHash.ToHexString())
					w.Write([]byte("\n"))
				}
				common.CheckErr(err)
				states = make([]*common3.TransferState, 0)
			} else {
				state := &common3.TransferState{
					From:  admin.Address,
					To:    toInfo.To,
					Value: toInfo.Amount.Uint64(),
				}
				states = append(states, state)
			}
		}
	}
}

func parseExcel(cf *config.Config) ([]*config.ToInfo, uint64) {
	f, err := excelize.OpenFile(cf.ExcelFile)
	common.CheckErr(err)
	// 获取 Sheet1 上所有单元格
	rows, err := f.GetRows("Sheet1")
	common.CheckErr(err)
	counter := 0
	var toInfos = make([]*config.ToInfo, 0)
	for i := 1; i < len(rows); i++ {
		counter++
		if rows[i][0] == "" {
			break
		}
		if !strings.HasPrefix(rows[i][0], "0x") {
			rows[i][0] = "0x" + rows[i][0]
		}
		to := web3.HexToAddress(rows[i][0])
		toInfos = append(toInfos, &config.ToInfo{
			To:     common2.Address(to),
			Amount: utils.ToIntByPrecise(rows[i][1], 9),
		})
	}
	sum := uint64(0)
	if !cf.Execute {
		for _, toInfo := range toInfos {
			sum += toInfo.Amount.Uint64()
			fmt.Println("to:", hex.EncodeToString(toInfo.To[:]), "amount:", toInfo.Amount.String())
		}
	}
	txNums := len(toInfos)/20*12/100 + 1
	return toInfos, sum + uint64(txNums)
}
