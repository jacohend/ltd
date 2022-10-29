#!/bin/bash
echo "Please Enter Network: "
read NETWORK || exit
export NETWORK=${NETWORK:-testnet}
echo "Please Enter Terminal Password: "
read PASS || exit
export PASS=${PASS:-changethis}
echo "Please Enter NAT flag (leave blank if unknown: "
read NAT || exit
export NAT=${NAT}
envsubst < ./config/ltd.service.template > ./config/ltd.service