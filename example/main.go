package main

import (
	"fmt"
	"github.com/howeyc/gopass"
	ontology_go_sdk "github.com/ontio/ontology-go-sdk"
	"github.com/ontio/ontology-go-sdk/oep4"
	common2 "github.com/ontio/ontology/common"
	"github.com/transerTools/common"
	"github.com/transerTools/utils"
	"github.com/urfave/cli/v2"
	"os"
)

func main() {
	app := &cli.App{
		Name:  "innoswap",
		Usage: "innoswap cli tool",
		Commands: []*cli.Command{
			removeLiquidityCmd,
			addLiquidityCmd,
		},
		Flags: []cli.Flag{
			WalletFileFlag,
			SignerFlag,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
	}
}

var removeLiquidityCmd = &cli.Command{
	Name:   "removeLiquidity",
	Action: removeLiquidity,
	Flags: []cli.Flag{
		AmountFlag,
		MinOntdFlag,
		MinTokensFlag,
		DeadlineFlag,
		WithdrawerFlag,
	},
}

func removeLiquidity(ctx *cli.Context) error {
	sdk, signer := initSdk(ctx)
	if sdk == nil || signer == nil {
		return nil
	}
	var contractAddress, _ = common2.AddressFromBase58("ANeRJJWVTpL77GSYwZqZK7gCZhAZ7u6osD")
	fmt.Println("contractAddress:", contractAddress.ToHexString())
	lpDecimal := getTokenDecimal(contractAddress, sdk)
	amtStr := ctx.String(AmountFlag.Name) // lp的数量
	amount := utils.ToIntByPrecise(amtStr, lpDecimal)
	minOntdStr := ctx.String(MinOntdFlag.Name)
	minOntd := utils.ToIntByPrecise(minOntdStr, 9)
	minTokensStr := ctx.String(MinTokensFlag.Name)
	tokenDecimal := getTokenDecimal(getTokenAddress(sdk, contractAddress), sdk)
	minTokens := utils.ToIntByPrecise(minTokensStr, tokenDecimal)
	deadline := ctx.Uint64(DeadlineFlag.Name)
	withdrawer := ctx.String(WithdrawerFlag.Name)
	wi, err := common2.AddressFromBase58(withdrawer)
	common.CheckErr(err)
	txHash, err := sdk.NeoVM.InvokeNeoVMContract(2500, 6000000, signer, contractAddress,
		[]interface{}{"removeLiquidity", []interface{}{amount, minOntd, minTokens, deadline, wi}})
	common.CheckErr(err)
	fmt.Println("removeLiquidity txHash:", txHash.ToHexString())
	return nil
}

func getTokenDecimal(adr common2.Address, sdk *ontology_go_sdk.OntologySdk) uint64 {
	o4 := oep4.NewOep4(adr, sdk)
	d, err := o4.Decimals()
	common.CheckErr(err)
	return d.Uint64()
}

func getTokenAddress(sdk *ontology_go_sdk.OntologySdk, contractAddress common2.Address) common2.Address {
	res, err := sdk.NeoVM.PreExecInvokeNeoVMContract(contractAddress, []interface{}{"tokenAddress", []interface{}{}})
	common.CheckErr(err)
	data, err := res.Result.ToByteArray()
	common.CheckErr(err)
	adr, err := common2.AddressParseFromBytes(data)
	common.CheckErr(err)
	return adr
}

var addLiquidityCmd = &cli.Command{
	Name:   "addLiquidity",
	Action: addLiquidity,
	Flags: []cli.Flag{
		MinLiquidateFlag,
		MaxTokensFlag,
		DeadlineFlag,
		DepositerFlag,
		DepositOntdAmtFlag,
	},
}

func addLiquidity(ctx *cli.Context) error {
	sdk, signer := initSdk(ctx)
	if sdk == nil || signer == nil {
		return nil
	}
	var contractAddress, _ = common2.AddressFromBase58("ANeRJJWVTpL77GSYwZqZK7gCZhAZ7u6osD")
	fmt.Println("contractAddress:", contractAddress.ToHexString())
	minLiquidateStr := ctx.String(MinLiquidateFlag.Name)
	lpDecimal := getTokenDecimal(contractAddress, sdk)
	minLiquidate := utils.ToIntByPrecise(minLiquidateStr, lpDecimal)
	tokenAddress := getTokenAddress(sdk, contractAddress)
	tokenDecimal := getTokenDecimal(tokenAddress, sdk)
	maxTokensStr := ctx.String(MaxTokensFlag.Name)
	maxToken := utils.ToIntByPrecise(maxTokensStr, tokenDecimal)
	deadline := ctx.Uint64(DeadlineFlag.Name)
	depositer := ctx.String(DepositerFlag.Name)
	de, err := common2.AddressFromBase58(depositer)
	common.CheckErr(err)
	depositOntdAmtStr := ctx.String(DepositOntdAmtFlag.Name)
	depositOntdAmt := utils.ToIntByPrecise(depositOntdAmtStr, 9)
	txHash, err := sdk.NeoVM.InvokeNeoVMContract(2500, 6000000, signer, contractAddress,
		[]interface{}{"addLiquidity", []interface{}{minLiquidate, maxToken, deadline, deadline, de, depositOntdAmt}})
	common.CheckErr(err)
	fmt.Println("addLiquidity txHash:", txHash.ToHexString())
	return nil
}

func initSdk(ctx *cli.Context) (*ontology_go_sdk.OntologySdk, *ontology_go_sdk.Account) {
	sdk := ontology_go_sdk.NewOntologySdk()
	sdk.NewRpcClient().SetAddress("http://dappnode2.ont.io:20336")

	walletFile := ctx.String(WalletFileFlag.Name)
	wall, err := sdk.OpenWallet(walletFile)
	common.CheckErr(err)
	address := ctx.String(SignerFlag.Name)
	pwd, err := GetPassword()
	if err != nil {
		fmt.Println("get password error")
		return nil, nil
	}
	admin, err := wall.GetAccountByAddress(address, pwd)
	common.CheckErr(err)
	fmt.Println("signer address:", admin.Address.ToBase58())
	return sdk, admin
}

// GetPassword gets password from user input
func GetPassword() ([]byte, error) {
	fmt.Printf("Password:")
	passwd, err := gopass.GetPasswd()
	if err != nil {
		return nil, err
	}
	return passwd, nil
}

var WalletFileFlag = &cli.StringFlag{
	Name:  "walletFile",
	Usage: "specify walletFile",
	Value: "",
}
var SignerFlag = &cli.StringFlag{
	Name:  "signer",
	Usage: "specify signer",
	Value: "",
}

var MinLiquidateFlag = &cli.StringFlag{
	Name:  "minLiquidate",
	Usage: "specify minLiquidate",
	Value: "",
}

var MaxTokensFlag = &cli.StringFlag{
	Name:  "maxTokens",
	Usage: "specify maxTokens",
	Value: "",
}

var DepositerFlag = &cli.StringFlag{
	Name:  "depositer",
	Usage: "specify depositer",
	Value: "",
}

var DepositOntdAmtFlag = &cli.StringFlag{
	Name:  "depositOntdAmt",
	Usage: "specify depositOntdAmt",
	Value: "",
}

var AmountFlag = &cli.StringFlag{
	Name:  "amount",
	Usage: "specify amount",
	Value: "",
}
var MinOntdFlag = &cli.StringFlag{
	Name:  "minOntd",
	Usage: "specify minOntd",
	Value: "",
}
var MinTokensFlag = &cli.StringFlag{
	Name:  "minTokens",
	Usage: "specify minTokens",
	Value: "",
}

var DeadlineFlag = &cli.Uint64Flag{
	Name:  "deadline",
	Usage: "specify deadline",
	Value: 0,
}
var WithdrawerFlag = &cli.StringFlag{
	Name:  "withdrawer",
	Usage: "specify withdrawer",
	Value: "",
}
