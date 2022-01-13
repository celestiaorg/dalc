FROM golang:1.17 as builder
RUN apt update && \
    apt install make git gcc

ENV HOME /dalc
COPY / ${HOME}
WORKDIR ${HOME}

RUN make build

# stage 2
FROM ubuntu

COPY celestia-light /root/.celestia-light
COPY docker/entrypoint.sh /

COPY --from=builder /dalc/dalc /

EXPOSE 26656 26657 1317 9090

ENTRYPOINT ["/entrypoint.sh"]
CMD ["dalc"]
