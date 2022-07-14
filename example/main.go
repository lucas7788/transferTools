package main

import (
	"fmt"
	"github.com/howeyc/gopass"
	ontology_go_sdk "github.com/ontio/ontology-go-sdk"
	"github.com/ontio/ontology-go-sdk/oep4"
	common2 "github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/transerTools/common"
	"github.com/transerTools/utils"
	"github.com/urfave/cli/v2"
	"math/big"
	"os"
	"time"
)

var ONTD = "2e0de81023ea6d32460244f29c57c84ce569e7b7"

func main() {
	app := &cli.App{
		Name:  "innoswap",
		Usage: "innoswap cli tool",
		Commands: []*cli.Command{
			removeLiquidityCmd,
			addLiquidityCmd,
			balanceOfCmd,
			generateWalletCmd,
			ontToTokenSwapInputCmd,
			ontToTokenSwapOutputCmd,
			tokenToOntSwapInputCmd,
			tokenToOntSwapOutputCmd,
		},
		Flags: []cli.Flag{
			WalletFileFlag,
			SignerFlag,
			PreFlag,
			TokenFlag,
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
	},
}

var ontToTokenSwapInputCmd = &cli.Command{
	Name:   "ontToTokenSwapInput",
	Action: ontToTokenSwapInput,
	Flags: []cli.Flag{
		OntdAmountFlag,
		MinTokensFlag,
		PreFlag,
	},
}

var ontToTokenSwapOutputCmd = &cli.Command{
	Name:   "ontToTokenSwapOutput",
	Action: ontToTokenSwapOutput,
	Flags: []cli.Flag{
		OntdAmountFlag,
		TokensBoughtFlag,
	},
}

var tokenToOntSwapInputCmd = &cli.Command{
	Name:   "tokenToOntSwapInput",
	Action: tokenToOntSwapInput,
	Flags: []cli.Flag{
		TokensSoldFlag,
		MinOntdFlag,
	},
}

var tokenToOntSwapOutputCmd = &cli.Command{
	Name:   "tokenToOntSwapOutput",
	Action: tokenToOntSwapOutput,
	Flags: []cli.Flag{
		OntdBoughtFlag,
		MaxTokensFlag,
	},
}

//var tokenToTokenSwapInputCmd = &cli.Command{
//	Name:   "tokenToTokenSwapInput",
//	Action: tokenToTokenSwapInput,
//	Flags: []cli.Flag{
//		OntdAmountFlag,
//		TokensBoughtFlag,
//	},
//}
//
//func tokenToTokenSwapInput(ctx *cli.Context) error {
//	sdk := initSdk()
//	if sdk == nil {
//		return nil
//	}
//	signer := initAccount(ctx, sdk)
//	if signer == nil {
//		return nil
//	}
//	contractAddress, _ := getContractAddress(ctx)
//	tokenDecimal := getTokenDecimal(getTokenAddress(sdk, contractAddress), sdk)
//	tokensSold := getAmount(ctx, TokensSoldFlag.Name, tokenDecimal)
//	minTokensBought := getAmount(ctx, MinTokensBoughtFlag.Name, tokenDecimal)
//
//	deadline := time.Now().Unix() + 1000
//	pre := ctx.Bool(PreFlag.Name)
//	if pre {
//		res, err := sdk.NeoVM.PreExecInvokeNeoVMContract(contractAddress,
//			[]interface{}{"tokenToOntSwapOutput",
//				[]interface{}{ontdBought, maxTokens, deadline, signer.Address}})
//		common.CheckErr(err)
//		arr, err := res.Result.ToInteger()
//		log.Infof("ontdBought: %s", utils.ToStringByPrecise(arr, 9))
//	} else {
//		txHash, err := sdk.NeoVM.InvokeNeoVMContract(2500, 6000000, signer, contractAddress,
//			[]interface{}{"tokenToOntSwapOutput",
//				[]interface{}{ontdBought, maxTokens, deadline, signer.Address}})
//		common.CheckErr(err)
//		log.Infof("tokenToOntSwapOutput txHash: %s", txHash.ToHexString())
//		sdk.WaitForGenerateBlock(time.Second*30, 1)
//		showLp(sdk, signer.Address, contractAddress)
//	}
//	return nil
//}

func getAmount(ctx *cli.Context, name string, decimals uint64) *big.Int {
	amtStr := ctx.String(name) // lp的数量
	return utils.ToIntByPrecise(amtStr, decimals)
}

func tokenToOntSwapOutput(ctx *cli.Context) error {
	sdk := initSdk()
	if sdk == nil {
		return nil
	}
	signer := initAccount(ctx, sdk)
	if signer == nil {
		return nil
	}
	contractAddress, _ := getContractAddress(ctx)
	tokenDecimal := getTokenDecimal(getTokenAddress(sdk, contractAddress), sdk)
	ontdBought := getAmount(ctx, OntdBoughtFlag.Name, 9)
	maxTokens := getAmount(ctx, MaxTokensFlag.Name, tokenDecimal)

	deadline := time.Now().Unix() + 1000
	pre := ctx.Bool(PreFlag.Name)
	approveOng(sdk, signer, contractAddress, maxTokens)
	if pre {
		res, err := sdk.NeoVM.PreExecInvokeNeoVMContract(contractAddress,
			[]interface{}{"tokenToOntSwapOutput",
				[]interface{}{ontdBought, maxTokens, deadline, signer.Address}})
		common.CheckErr(err)
		arr, err := res.Result.ToInteger()
		log.Infof("ontdBought: %s", utils.ToStringByPrecise(arr, 9))
	} else {
		txHash, err := sdk.NeoVM.InvokeNeoVMContract(2500, 6000000, signer, contractAddress,
			[]interface{}{"tokenToOntSwapOutput",
				[]interface{}{ontdBought, maxTokens, deadline, signer.Address}})
		common.CheckErr(err)
		log.Infof("tokenToOntSwapOutput txHash: %s", txHash.ToHexString())
		sdk.WaitForGenerateBlock(time.Second*30, 1)
		showLp(sdk, signer.Address, contractAddress)
	}
	return nil
}

func tokenToOntSwapInput(ctx *cli.Context) error {
	sdk := initSdk()
	if sdk == nil {
		return nil
	}
	signer := initAccount(ctx, sdk)
	if signer == nil {
		return nil
	}
	contractAddress, _ := getContractAddress(ctx)
	tokenDecimal := getTokenDecimal(getTokenAddress(sdk, contractAddress), sdk)
	amtStr := ctx.String(TokensSoldFlag.Name) // lp的数量
	tokenSoldAmount := utils.ToIntByPrecise(amtStr, tokenDecimal)
	minOntdStr := ctx.String(MinOntdFlag.Name)
	minOntd := utils.ToIntByPrecise(minOntdStr, 9)

	deadline := time.Now().Unix() + 1000
	pre := ctx.Bool(PreFlag.Name)
	approveOng(sdk, signer, contractAddress, tokenSoldAmount)
	if pre {
		res, err := sdk.NeoVM.PreExecInvokeNeoVMContract(contractAddress,
			[]interface{}{"tokenToOntSwapInput",
				[]interface{}{tokenSoldAmount, minOntd, deadline, signer.Address}})
		common.CheckErr(err)
		arr, err := res.Result.ToInteger()
		log.Infof("ontdBought: %s", utils.ToStringByPrecise(arr, 9))
	} else {
		txHash, err := sdk.NeoVM.InvokeNeoVMContract(2500, 6000000, signer, contractAddress,
			[]interface{}{"tokenToOntSwapInput",
				[]interface{}{tokenSoldAmount, minOntd, deadline, signer.Address}})
		common.CheckErr(err)
		log.Infof("tokenToOntSwapInput txHash: %s", txHash.ToHexString())
		sdk.WaitForGenerateBlock(time.Second*30, 1)
		showLp(sdk, signer.Address, contractAddress)
	}
	return nil
}

func getContractAddress(ctx *cli.Context) (common2.Address, string) {
	name := ctx.String(TokenFlag.Name)
	if name == "WING" {
		var contractAddress, _ = common2.AddressFromBase58("ANeRJJWVTpL77GSYwZqZK7gCZhAZ7u6osD")
		fmt.Println("contractAddress:", contractAddress.ToHexString())
		return contractAddress, name
	} else if name == "ONG" {
		var contractAddress, _ = common2.AddressFromHexString("969c9028bb048c4e1c2120129bc488bf8b5ad41b")
		fmt.Println("contractAddress:", contractAddress.ToHexString())
		return contractAddress, name
	} else {
		panic(" not support token " + name)
	}
}

func ontToTokenSwapOutput(ctx *cli.Context) error {
	sdk := initSdk()
	if sdk == nil {
		return nil
	}
	signer := initAccount(ctx, sdk)
	if signer == nil {
		return nil
	}

	contractAddress, name := getContractAddress(ctx)
	amtStr := ctx.String(OntdAmountFlag.Name) // lp的数量
	ontdAmount := utils.ToIntByPrecise(amtStr, 9)
	tokensBoughtStr := ctx.String(TokensBoughtFlag.Name)
	tokenDecimal := getTokenDecimal(getTokenAddress(sdk, contractAddress), sdk)
	tokensBought := utils.ToIntByPrecise(tokensBoughtStr, tokenDecimal)
	deadline := time.Now().Unix() + 1000
	pre := ctx.Bool(PreFlag.Name)
	approveOntd(sdk, signer, contractAddress, ontdAmount)
	if pre {
		res, err := sdk.NeoVM.PreExecInvokeNeoVMContract(contractAddress,
			[]interface{}{"ontToTokenSwapOutput", []interface{}{tokensBought, deadline, signer.Address, ontdAmount}})
		common.CheckErr(err)
		arr, err := res.Result.ToArray()
		common.CheckErr(err)
		b1, err := arr[0].ToInteger()
		b2, err := arr[0].ToInteger()
		log.Infof("ontdAmount: %s, %s tokenAmount: %s", utils.ToStringByPrecise(b1, 9), name, utils.ToStringByPrecise(b2, tokenDecimal))
	} else {
		txHash, err := sdk.NeoVM.InvokeNeoVMContract(2500, 6000000, signer, contractAddress,
			[]interface{}{"ontToTokenSwapOutput", []interface{}{tokensBought, deadline, signer.Address, ontdAmount}})
		common.CheckErr(err)
		log.Infof("ontToTokenSwapOutput txHash: %s", txHash.ToHexString())
		sdk.WaitForGenerateBlock(time.Second*30, 1)
		showLp(sdk, signer.Address, contractAddress)
	}
	return nil
}

func ontToTokenSwapInput(ctx *cli.Context) error {
	sdk := initSdk()
	if sdk == nil {
		return nil
	}
	signer := initAccount(ctx, sdk)
	if signer == nil {
		return nil
	}

	contractAddress, _ := getContractAddress(ctx)
	amtStr := ctx.String(OntdAmountFlag.Name) // lp的数量
	ontdAmount := utils.ToIntByPrecise(amtStr, 9)
	minTokensStr := ctx.String(MinTokensFlag.Name)
	tokenDecimal := getTokenDecimal(getTokenAddress(sdk, contractAddress), sdk)
	minTokens := utils.ToIntByPrecise(minTokensStr, tokenDecimal)
	deadline := time.Now().Unix() + 1000
	pre := ctx.Bool(PreFlag.Name)
	approveOntd(sdk, signer, contractAddress, ontdAmount)
	if pre {
		res, err := sdk.NeoVM.PreExecInvokeNeoVMContract(contractAddress,
			[]interface{}{"ontToTokenSwapInput", []interface{}{minTokens, deadline, signer.Address, ontdAmount}})
		common.CheckErr(err)
		arr, err := res.Result.ToInteger()
		common.CheckErr(err)
		log.Infof("tokenBought: %s", utils.ToStringByPrecise(arr, 9))
	} else {
		fmt.Println(minTokens, ontdAmount)
		txHash, err := sdk.NeoVM.InvokeNeoVMContract(2500, 6000000, signer, contractAddress,
			[]interface{}{"ontToTokenSwapInput", []interface{}{minTokens, deadline, signer.Address, ontdAmount}})
		common.CheckErr(err)
		log.Infof("ontToTokenSwapInput txHash: %s", txHash.ToHexString())
		sdk.WaitForGenerateBlock(time.Second*30, 1)
		showLp(sdk, signer.Address, contractAddress)
	}
	return nil
}

func approveOntd(sdk *ontology_go_sdk.OntologySdk, signer *ontology_go_sdk.Account,
	spender common2.Address, amt *big.Int) {
	ontdAddr, _ := common2.AddressFromHexString(ONTD)
	all, err := sdk.NeoVM.PreExecInvokeNeoVMContract(ontdAddr,
		[]interface{}{"allowance", []interface{}{signer.Address, spender}})
	if err != nil {
		panic(err)
	}
	al, err := all.Result.ToInteger()
	common.CheckErr(err)
	if al.Cmp(amt) <= 0 {
		ontd := oep4.NewOep4(ontdAddr, sdk)
		bal, err := ontd.BalanceOf(signer.Address)
		common.CheckErr(err)
		txHash, err := ontd.Approve(signer, spender, bal, 2500, 30000)
		common.CheckErr(err)
		fmt.Println("approveTxHash:", txHash.ToHexString())
		sdk.WaitForGenerateBlock(time.Second*30, 1)
	}
}

func approveOng(sdk *ontology_go_sdk.OntologySdk, signer *ontology_go_sdk.Account,
	spender common2.Address, amt *big.Int) {
	all, err := sdk.Native.Ong.Allowance(signer.Address, spender)
	if err != nil {
		panic(err)
	}
	if all < amt.Uint64() {
		bal, err := sdk.Native.Ong.BalanceOf(signer.Address)
		common.CheckErr(err)
		txHash, err := sdk.Native.Ong.Approve(2500, 30000, signer, spender, bal)
		common.CheckErr(err)
		fmt.Println("approveTxHash:", txHash.ToHexString())
		sdk.WaitForGenerateBlock(time.Second*30, 1)
	}
}

func removeLiquidity(ctx *cli.Context) error {
	sdk := initSdk()
	if sdk == nil {
		return nil
	}
	signer := initAccount(ctx, sdk)
	if signer == nil {
		return nil
	}

	contractAddress, name := getContractAddress(ctx)
	lpDecimal := getTokenDecimal(contractAddress, sdk)
	amtStr := ctx.String(AmountFlag.Name) // lp的数量
	amount := utils.ToIntByPrecise(amtStr, lpDecimal)
	minOntdStr := ctx.String(MinOntdFlag.Name)
	minOntd := utils.ToIntByPrecise(minOntdStr, 9)
	minTokensStr := ctx.String(MinTokensFlag.Name)
	tokenDecimal := getTokenDecimal(getTokenAddress(sdk, contractAddress), sdk)
	minTokens := utils.ToIntByPrecise(minTokensStr, tokenDecimal)
	deadline := time.Now().Unix() + 1000
	pre := ctx.Bool(PreFlag.Name)
	if pre {
		res, err := sdk.NeoVM.PreExecInvokeNeoVMContract(contractAddress,
			[]interface{}{"removeLiquidity", []interface{}{amount, minOntd, minTokens, deadline, signer.Address}})
		common.CheckErr(err)
		arr, err := res.Result.ToArray()
		common.CheckErr(err)
		b1, err := arr[0].ToInteger()
		b2, err := arr[0].ToInteger()
		log.Infof("ontdAmount: %s, %s tokenAmount: %s", utils.ToStringByPrecise(b1, 9), name, utils.ToStringByPrecise(b2, tokenDecimal))
	} else {
		fmt.Println(amount, minOntd, minTokens, deadline, signer.Address.ToBase58())
		txHash, err := sdk.NeoVM.InvokeNeoVMContract(2500, 6000000, signer, contractAddress,
			[]interface{}{"removeLiquidity", []interface{}{amount, minOntd, minTokens,
				deadline, signer.Address}})
		common.CheckErr(err)
		log.Infof("removeLiquidity txHash: %s", txHash.ToHexString())
		sdk.WaitForGenerateBlock(time.Second*30, 2)
		showLp(sdk, signer.Address, contractAddress)
	}
	return nil
}

func getTokenDecimal(adr common2.Address, sdk *ontology_go_sdk.OntologySdk) uint64 {
	if adr == ontology_go_sdk.ONG_CONTRACT_ADDRESS {
		return 9
	} else {
		fmt.Println("adr:", adr.ToHexString())
		o4 := oep4.NewOep4(adr, sdk)
		d, err := o4.Decimals()
		common.CheckErr(err)
		fmt.Println("adr:", adr.ToHexString(), "decimals:", d.Uint64())
		return d.Uint64()
	}
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
		DepositOntdAmtFlag,
	},
}

var createLpCmd = &cli.Command{
	Name:   "createLp",
	Action: createLp,
	Flags: []cli.Flag{
		MinLiquidateFlag,
		MaxTokensFlag,
		DepositOntdAmtFlag,
	},
}

func createLp(ctx *cli.Context) error {
	return nil
}

var balanceOfCmd = &cli.Command{
	Name:   "balanceOf",
	Action: balanceOf,
	Flags: []cli.Flag{
		AddressFlag,
	},
}
var generateWalletCmd = &cli.Command{
	Name:   "genWallet",
	Action: genWallet,
	Flags: []cli.Flag{
		PwdFlag,
	},
}

func genWallet(ctx *cli.Context) error {
	sdk := initSdk()
	if sdk == nil {
		return nil
	}
	pwd := ctx.String(PwdFlag.Name)
	var wall *ontology_go_sdk.Wallet
	var err error
	if common2.FileExisted("wallet.dat") {
		wall, err = sdk.OpenWallet("wallet.dat")
	} else {
		wall, err = sdk.CreateWallet("wallet.dat")
	}
	common.CheckErr(err)
	acc, err := wall.NewDefaultSettingAccount([]byte(pwd))
	common.CheckErr(err)
	log.Infof("user address: %s", acc.Address.ToBase58())
	wall.Save()
	return nil
}

func balanceOf(ctx *cli.Context) error {
	sdk := initSdk()
	if sdk == nil {
		return nil
	}
	addr := ctx.String(AddressFlag.Name)
	user, err := common2.AddressFromBase58(addr)
	common.CheckErr(err)
	lpAddr, _ := getContractAddress(ctx)
	showLp(sdk, user, lpAddr)
	return nil
}

func showLp(sdk *ontology_go_sdk.OntologySdk, user, lpAddr common2.Address) {
	lp := oep4.NewOep4(lpAddr, sdk)
	bal, err := lp.BalanceOf(user)
	common.CheckErr(err)
	log.Infof("user address: %s, lp balance: %s", user.ToBase58(), utils.ToStringByPrecise(bal, 9))
}

func addLiquidity(ctx *cli.Context) error {
	sdk := initSdk()
	if sdk == nil {
		return nil
	}
	signer := initAccount(ctx, sdk)
	if signer == nil {
		return nil
	}
	var contractAddress, _ = getContractAddress(ctx)
	minLiquidateStr := ctx.String(MinLiquidateFlag.Name)
	lpDecimal := getTokenDecimal(contractAddress, sdk)
	minLiquidate := utils.ToIntByPrecise(minLiquidateStr, lpDecimal)
	tokenAddress := getTokenAddress(sdk, contractAddress)
	tokenDecimal := getTokenDecimal(tokenAddress, sdk)
	maxTokensStr := ctx.String(MaxTokensFlag.Name)
	maxToken := utils.ToIntByPrecise(maxTokensStr, tokenDecimal)
	depositOntdAmtStr := ctx.String(DepositOntdAmtFlag.Name)
	depositOntdAmt := utils.ToIntByPrecise(depositOntdAmtStr, 9)

	var approveTxHash common2.Uint256
	if tokenAddress == ontology_go_sdk.ONG_CONTRACT_ADDRESS {
		all, err := sdk.Native.Ong.Allowance(signer.Address, contractAddress)
		common.CheckErr(err)
		if maxToken.Uint64() > all {
			bal, err := sdk.Native.Ong.BalanceOf(signer.Address)
			common.CheckErr(err)
			approveTxHash, err = sdk.Native.Ong.Approve(2500, 30000, signer, contractAddress,
				bal)
			common.CheckErr(err)
			log.Infof("approveTxHash: %s", approveTxHash.ToHexString())
			sdk.WaitForGenerateBlock(time.Second*30, 1)
		}
	} else {
		token := oep4.NewOep4(tokenAddress, sdk)
		all, err := sdk.NeoVM.PreExecInvokeNeoVMContract(tokenAddress, []interface{}{"allowance",
			[]interface{}{signer.Address, contractAddress}})
		common.CheckErr(err)
		ball, err := all.Result.ToInteger()
		common.CheckErr(err)
		log.Infof("allowance: %s", utils.ToStringByPrecise(ball, tokenDecimal))
		if ball.Cmp(maxToken) < 0 {
			bal, err := token.BalanceOf(signer.Address)
			common.CheckErr(err)
			approveTxHash, err = token.Approve(signer, contractAddress, bal,
				2500, 20000)
			common.CheckErr(err)
			log.Infof("approveTxHash: %s", approveTxHash.ToHexString())
			sdk.WaitForGenerateBlock(time.Second*30, 1)
		}
	}
	pre := ctx.Bool(PreFlag.Name)
	deadline := time.Now().Unix() + 1000
	if pre {
		res, err := sdk.NeoVM.PreExecInvokeNeoVMContract(contractAddress,
			[]interface{}{"addLiquidity", []interface{}{minLiquidate, maxToken, deadline,
				signer.Address, depositOntdAmt}})
		common.CheckErr(err)
		b, err := res.Result.ToInteger()
		common.CheckErr(err)
		log.Infof("liquidityMinted: %s", utils.ToStringByPrecise(b, lpDecimal))
	} else {
		fmt.Println(minLiquidate.String(), maxToken.String(), deadline)
		txHash, err := sdk.NeoVM.InvokeNeoVMContract(2500, 6000000, signer, contractAddress,
			[]interface{}{"addLiquidity", []interface{}{minLiquidate, maxToken, deadline,
				signer.Address, depositOntdAmt}})
		common.CheckErr(err)
		log.Infof("addLiquidity txHash: %s", txHash.ToHexString())
		sdk.WaitForGenerateBlock(time.Second*30, 2)
		showLp(sdk, signer.Address, contractAddress)
	}
	return nil
}

func initSdk() *ontology_go_sdk.OntologySdk {
	sdk := ontology_go_sdk.NewOntologySdk()
	//sdk.NewRpcClient().SetAddress("http://dappnode2.ont.io:20336")
	sdk.NewRpcClient().SetAddress("http://polaris2.ont.io:20336")
	return sdk
}

func initAccount(ctx *cli.Context, sdk *ontology_go_sdk.OntologySdk) *ontology_go_sdk.Account {
	walletFile := ctx.String(WalletFileFlag.Name)
	if walletFile == "" {
		panic("wallet file is nil")
	}
	wall, err := sdk.OpenWallet(walletFile)
	common.CheckErr(err)
	address := ctx.String(SignerFlag.Name)
	if address == "" {
		panic("signer is nil")
	}
	pwd, err := GetPassword()
	if err != nil {
		fmt.Println("get password error")
		return nil
	}
	admin, err := wall.GetAccountByAddress(address, pwd)
	common.CheckErr(err)
	fmt.Println("signer address:", admin.Address.ToBase58())
	return admin
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

var TokenFlag = &cli.StringFlag{
	Name:  "token",
	Usage: "specify token name",
	Value: "",
}

var PreFlag = &cli.BoolFlag{
	Name:  "pre",
	Usage: "specify pre",
	Value: false,
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

var AddressFlag = &cli.StringFlag{
	Name:  "address",
	Usage: "specify address",
	Value: "",
}

var PwdFlag = &cli.StringFlag{
	Name:  "pwd",
	Usage: "specify password",
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
var OntdAmountFlag = &cli.StringFlag{
	Name:  "ontd-amount",
	Usage: "specify ontd amount",
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
var TokensBoughtFlag = &cli.StringFlag{
	Name:  "tokensBought",
	Usage: "specify tokensBought",
	Value: "",
}

var MinTokensBoughtFlag = &cli.StringFlag{
	Name:  "minTokensBought",
	Usage: "specify minTokensBought",
	Value: "",
}

var TokensSoldFlag = &cli.StringFlag{
	Name:  "tokensSold",
	Usage: "specify tokensSold",
	Value: "",
}

var OntdBoughtFlag = &cli.StringFlag{
	Name:  "ontdBought",
	Usage: "specify ontdBought",
	Value: "",
}
