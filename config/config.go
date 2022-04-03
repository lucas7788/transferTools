package config

import (
	"encoding/json"
	common2 "github.com/ontio/ontology/common"
	"github.com/transerTools/common"
	"io/ioutil"
	"math/big"
)

type Config struct {
	WalletFile      string `json:"walletFile"`
	Password        string `json:"password"`
	Address         string `json:"address"`
	ExcelFile       string `json:"excelFile"`
	ContractAddress string `json:"contractAddress"`
	RpcUrl          string `json:"rpcUrl"`
	GasPrice        uint64 `json:"gasPrice"`
	GasLimit        uint64 `json:"gasLimit"`
	Execute         bool   `json:"execute"`
}

func (this *Config) GetContractAddress() common2.Address {
	if this.ContractAddress == "" {
		panic("this.ContractAddress is nil")
	}
	adr, err := common2.AddressFromHexString(this.ContractAddress)
	common.CheckErr(err)
	return adr
}

func ParseConfig() *Config {
	data, err := ioutil.ReadFile("./config.json")
	common.CheckErr(err)
	var confg Config
	err = json.Unmarshal(data, &confg)
	common.CheckErr(err)
	return &confg
}

type ToInfo struct {
	To     common2.Address
	Amount *big.Int
}
