#!/bin/bash

# Set variables
capp=celestia-appd
CHAIN_ID="ephemeral"
KEY_TYPE="test"

if [ $# != 1 ]; then
	echo -e "Usage:\n$0 <NODE_NAME>"
	exit 1
fi
NODE_NAME=$1

# Creating the account for validator #1
$capp --home $NODE_NAME keys add $NODE_NAME --keyring-backend=$KEY_TYPE
node_addr=$($capp --home $NODE_NAME keys show $NODE_NAME -a --keyring-backend $KEY_TYPE)

echo $node_addr