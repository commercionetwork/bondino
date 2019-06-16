# Bondino
This repo contains the code that represents the entry for the 2019 Berlin Cosmos Hackatom from Commercio.network. 

Bodino represents a pool-backed system that allows user to open Credit Debt Positions using any sort of FT as well 
as NFT.  
CDPs can be opened even when the price of the token is not established yet.  
Once a price is correctly known, the user that has opened the CDP will automatically receive atoms taken from a pool of 
investors, based on certain conditions (i.e. the warranty ratio). 

## Credits
Our work is a fork of [Kava](https://github.com/Kava-Labs/kava-devnet), to which we added a new module called `pool`.

We also extended the original Kava code in order to support NFTs as well as FTs, even when the prices can not be 
retrieved immediately but will later be inserted from an external oracle.

## Usage
### Pool
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