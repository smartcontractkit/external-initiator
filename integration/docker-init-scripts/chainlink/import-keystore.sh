#!/bin/bash

echo "** Importing default key 0x9ca9d2d5e04012c9ed24c0e513c9bfaa4a2dd77f"
chainlink node import /run/secrets/keystore

echo "** Running node"
chainlink node start -d -p /run/secrets/node_password -a /run/secrets/apicredentials
