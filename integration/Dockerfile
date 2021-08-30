
FROM public.ecr.aws/chainlink/chainlink:0.10.13

COPY ./docker-init-scripts/chainlink/import-keystore.sh ./

ENTRYPOINT ["./import-keystore.sh"]
