package main

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"github.com/laizy/web3"
	ontology_go_sdk "github.com/ontio/ontology-go-sdk"
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
		states := make([]*ont.State, 0)
		end := len(toInfos)
		for k, toInfo := range toInfos {
			if len(states) >= 20 || k == end-1 {
				txHash, err := sdk.Native.Ong.MultiTransfer(cf.GasPrice, cf.GasLimit, states, admin)
				if cf.Execute {
					w.WriteString(txHash.ToHexString())
					w.Write([]byte("\n"))
				}
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
	}
}

func parseExcel(cf *config.Config) ([]*config.ToInfo, uint64) {
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
			Amount: utils.ToIntByPrecise(rows[i][1], 9),
		})
	}
	sum := uint64(0)
	if !cf.Execute {
		for _, toInfo := range toInfos {
			sum += toInfo.Amount.Uint64()
			if isBase58 {
				fmt.Println("to:", toInfo.To.ToBase58(), "amount:", toInfo.Amount.String())
			} else {
				fmt.Println("to:", "0x"+hex.EncodeToString(toInfo.To[:]), "amount:", toInfo.Amount.String())
			}
		}
	}
	fee := (len(toInfos)/20)*12/100 + 1
	if !cf.Execute {
		fmt.Println("estimate tx fee is: ", fee)
	}
	return toInfos, sum + uint64(fee)
}
