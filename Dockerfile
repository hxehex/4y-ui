FROM --platform=$BUILDPLATFORM tonistiigi/xx:1.3.0 AS xx

FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS builder
COPY --from=xx / /
WORKDIR /app

ARG TARGETPLATFORM
ARG TARGETARCH

RUN apk --no-cache --update add \
  clang \
  lld \
  curl \
  unzip

RUN xx-apk --no-cache --update add \
  build-base \
  gcc \
  musl-dev

COPY . .

ENV CGO_ENABLED=1
ENV CGO_CFLAGS="-D_LARGEFILE64_SOURCE"

RUN xx-go build -ldflags "-w -s" -o build/4y-ui main.go

RUN ./DockerInit.sh "$TARGETARCH"

FROM alpine
ENV TZ=Asia/Tehran
WORKDIR /app

RUN apk add --no-cache --update \
  ca-certificates \
  tzdata \
  fail2ban \
  bash \
  curl

COPY --from=builder /app/build/ /app/
COPY --from=builder /app/DockerEntrypoint.sh /app/
COPY --from=builder /app/4y-ui.sh /usr/bin/4y-ui

RUN rm -f /etc/fail2ban/jail.d/alpine-ssh.conf \
  && cp /etc/fail2ban/jail.conf /etc/fail2ban/jail.local \
  && sed -i "s/^\[ssh\]$/&\nenabled = false/" /etc/fail2ban/jail.local \
  && sed -i "s/^\[sshd\]$/&\nenabled = false/" /etc/fail2ban/jail.local \
  && sed -i "s/#allowipv6 = auto/allowipv6 = auto/g" /etc/fail2ban/fail2ban.conf

RUN chmod +x \
  /app/DockerEntrypoint.sh \
  /app/4y-ui \
  /usr/bin/4y-ui

ENV XUI_ENABLE_FAIL2BAN="true"
EXPOSE 2053
VOLUME [ "/etc/4y-ui" ]
CMD [ "./4y-ui" ]
ENTRYPOINT [ "/app/DockerEntrypoint.sh" ]