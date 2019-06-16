# ❄️ Bondino ❄️
This repo contains the code that represents the entry for the 2019 Berlin Cosmos Hackatom from [Commercio.network](https://commercio.network). 

Bodino represents a pool-backed system that allows user to open Credit Debt Positions using any sort of FT as well 
as NFT.  
CDPs can be opened even when the price of the token is not established yet.  
Once a price is correctly known, the user that has opened the CDP will automatically receive atoms taken from a pool of 
investors, based on certain conditions (i.e. the warranty ratio). 

## 📝 Credits 📝
Our work is a fork of [Kava](https://github.com/Kava-Labs/kava-devnet), to which we added a new module called `pool`.

We also extended the original Kava code in order to support NFTs as well as FTs, even when the prices can not be 
retrieved immediately but will later be inserted from an external oracle.

Many thanks goes to Ruaridh ([@rhuairahrighairidh](https://github.com/rhuairahrighairidh)) that allowed us to use its code. We couldn't have done it without it! 💯

## ⚠️ Warning ⚠️
This repository in under **heavy** development and **should not** be used in production whatsover. Please not that we will strive to implement all the missing features as soon as possible, and fix all the bugs that are present. 

We also aim to re-implement the currently not-working tests, and create other ones to further improve its stability. 

If you are willing to help us, please open an issue for any bug you find, or create a pull request to speed up the overall development of the features 💪

## 📜 Usage 📜
### Pool

> Allows to deposit and withdraw a given token amount into/from the liquidity pool .

Deposit a given amount 
```bash
kavacli tx pool deposit [amount] --from <key_name>

E.g. kavacli tx pool deposit 1000uatom --from jack
``` 

Withdraw a deposited amount
```bash
kavacli tx pool withdraw [amount] -- from <key_name>

E.g. kavacli tx pool withdraw 500uatom --from jack
```

See the current deposited amount for a given user
```bash
kavacli query pool get-funds [address]

E.g. kavacli query pool get-funds $(kavacli keys show jack --address) 
```

See all the deposited funds 
```bash
kavacli query pool funds 
``` 

### Auction
> Allows to close a collateralized debt position (CDP). 

🔨 **WIP** 🔨

### Collateralized Debt Position
> Allows to open a collateralized debt position (CDP) and later edit it.

🔨 **WIP** 🔨

### Liquidator
> Allows to automatically liquidate a CDP when some criteria are met (change in the liquidity/collaterals value)

🔨 **WIP** 🔨

### Price feed
> Allows to fetch the prices of non-fungible and fungible tokens contacting external oracles

🔨 **WIP** 🔨
