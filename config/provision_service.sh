#!/bin/bash
echo "Please Enter Network: " && (read NETWORK || exit)
export NETWORK=${NETWORK:-testnet}
echo "Please Enter Terminal Password: "
read PASS || exit
export PASS=${PASS:-changethis}
envsubst < ./config/ltd.service.template > ./config/ltd.service