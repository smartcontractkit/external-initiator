# TODO: smartcontract/chainlink:latest is BLOCKED by: https://www.pivotaltracker.com/n/projects/2441917/stories/174807792
FROM smartcontract/chainlink:0.8.12

COPY ./docker-init-scripts/chainlink/import-keystore.sh ./

ENTRYPOINT ["./import-keystore.sh"]
