#!/bin/sh
set -o errexit -o nounset

CHAINID="test"

# Build genesis file incl account for passed address
coins="10000000000000000stake,1000000000000000tia"
celestia-appd init $CHAINID --chain-id $CHAINID

# create genesis account keys
celestia-appd keys add validator --keyring-backend test 
celestia-appd keys add user1 --keyring-backend test 
celestia-appd keys add user2 --keyring-backend test 

# create genesis accounts 
celestia-appd add-genesis-account $(celestia-appd keys show validator -a --keyring-backend test) $coins 
celestia-appd add-genesis-account user1 $coins --keyring-backend test 
celestia-appd add-genesis-account user2 $coins --keyring-backend test 

# create the first validator 
celestia-appd gentx validator 5000000000stake --keyring-backend test --chain-id $CHAINID 

# finalize the genesis file 
celestia-appd collect-gentxs 

# Set proper defaults and change ports
sed -i 's#"tcp://127.0.0.1:26657"#"tcp://0.0.0.0:26657"#g' ~/.celestia-app/config/config.toml
sed -i 's/timeout_commit = "5s"/timeout_commit = "1s"/g' ~/.celestia-app/config/config.toml
sed -i 's/timeout_propose = "3s"/timeout_propose = "1s"/g' ~/.celestia-app/config/config.toml
sed -i 's/index_all_keys = false/index_all_keys = true/g' ~/.celestia-app/config/config.toml

# Start the celestia-app
echo "running celestia-app"
celestia-appd start 