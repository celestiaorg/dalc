FROM golang:1.17 as builder
RUN apt update && \
    apt install make git gcc

ENV HOME /dalc
COPY / ${HOME}
WORKDIR ${HOME}

RUN make build

# stage 2
FROM ubuntu

COPY docker/entrypoint.sh /root/entrypoint.sh

# Copy in the binary
COPY --from=builder /dalc/celestia /root/celestia

WORKDIR /root

EXPOSE 4200

ENTRYPOINT ["/root/entrypoint.sh"]
CMD ["celestia"]
