FROM registry.access.redhat.com/ubi8/go-toolset:1.18 AS golang

WORKDIR /tmp/valigator
ADD cmd ./cmd
ADD go.mod .
ADD go.sum .
WORKDIR /tmp/valigator/cmd

USER 0
RUN GOARCH=arm64 go build -v -ldflags '-s -w' -o "out/valigator" .
USER 1001

FROM registry.access.redhat.com/ubi8/nodejs-16-minimal

RUN npm install -g @stoplight/spectral@5.9.2

ENV NODE_ENV production

ADD --chown=1001:0 docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh
ENTRYPOINT [ "/docker-entrypoint.sh" ]

COPY --from=golang --chown=1001:0 /tmp/valigator/cmd/out/valigator /usr/local/bin/valigator

USER 0
RUN chmod +x /usr/local/bin/valigator
USER 1001

COPY --chown=1001:1001  valigator.json $HOME/valigator.json

CMD [ "valigator" ]