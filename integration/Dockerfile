
FROM smartcontract/chainlink:latest

COPY ./docker-init-scripts/chainlink/import-keystore.sh ./

ENTRYPOINT ["./import-keystore.sh"]
