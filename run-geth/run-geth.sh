#!/bin/bash

if [ ! -d "data" ]; then
    mkdir data
fi

if [ ! -d "data/geth" ]; then
    geth --datadir ./data init ./genesis.json
fi

geth \
    --dev \
    --dev.period 5 \
    --datadir data \
    --http \
    --http.port 8545 \
    --http.addr 0.0.0.0 \
    --http.corsdomain '*' \
    --http.api eth,net,web3,personal,debug \
    --ws \
    --ws.origins '*' \
    --ws.addr 0.0.0.0 \
    --ws.port 8546 \
    --ws.api eth,net,web3,personal,debug
