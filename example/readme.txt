
添加ONG和ontd流动性
 ./main --signer AUfgnmUz6tTj7anRYSjz8kdTGHet816Vzn --token ONG --walletFile wallet.dat addLiquidity --minLiquidate 1 --maxTokens 2 --depositOntdAmt 1

minLiquidate 表示得到的流动性lp的最小数量
maxTokens 表示需要的WING的最大数量
depositOntdAmt 表示要充的ontd的数量


删除流动性
 ./main --walletFile ../wallet.dat --token ONG --signer AUfgnmUz6tTj7anRYSjz8kdTGHet816Vzn removeLiquidity --amount 1 --minOntd 0.001 --minTokens 0.001

amount 表示要删除的流动性的数量
minOntd 表示接收到的ontd的最小数量
minTokens 表示接收到的WING的最小数量


查询lp的余额
./main balanceOf --token ONG --address AY18Azi35MLDpCUeKeTtSmNPZH38HaBJgH


ontd 换ONG

方法1  建议用该方法
 ./main --walletFile wallet.dat --token ONG --signer AUfgnmUz6tTj7anRYSjz8kdTGHet816Vzn ontToTokenSwapInput --ontd-amount 1 --minTokens 0.001

方法2
 ./main --walletFile wallet.dat --token ONG --signer AUfgnmUz6tTj7anRYSjz8kdTGHet816Vzn ontToTokenSwapOutput --ontd-amount 1 --tokensBought 0.001

ONG 换 ontd

方法1  建议用该方法

 ./main --walletFile wallet.dat --token ONG --signer AUfgnmUz6tTj7anRYSjz8kdTGHet816Vzn tokenToOntSwapInput --tokensSold 1 --minOntd 0.001


方法2
 ./main --walletFile wallet.dat --token ONG --signer AUfgnmUz6tTj7anRYSjz8kdTGHet816Vzn tokenToOntSwapOutput --ontdBought 1 --maxTokens 2
