ARG WODBY_BASE_IMAGE
FROM ${WODBY_BASE_IMAGE} AS builder

ARG COPY_FROM
COPY --chown=wodby:wodby ${COPY_FROM}/go.mod /usr/src/app/
RUN go mod download

COPY --chown=wodby:wodby ${COPY_FROM} /usr/src/app
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /home/wodby/go/bin/app .

FROM ${WODBY_BASE_IMAGE}
COPY --from=builder --chown=wodby:wodby /home/wodby/go/bin/app /home/wodby/go/bin/app

CMD ["app"]
