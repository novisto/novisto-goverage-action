FROM golang:1.22 AS BUILD

RUN apt update \
    && apt install -y git \
    && apt clean

WORKDIR /go/src/app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags '-extldflags "-static"' -tags timetzdata -o goverage

FROM gcr.io/distroless/static-debian11:nonroot

COPY --from=BUILD /go/src/app/goverage /goverage

ENTRYPOINT ["/goverage"]
