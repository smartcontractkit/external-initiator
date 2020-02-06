FROM smartcontract/chainlink:develop

COPY ./docker-init-scripts/chainlink/import-keystore.sh ./

ENTRYPOINT ["./import-keystore.sh"]
