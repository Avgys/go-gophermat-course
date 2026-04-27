# Prebuilt Practicum binary: place accrual_linux_amd64 next to this file.
# Build from repository root:
#   docker build -f cmd/accrual/Dockerfile .

FROM debian:bookworm-slim

RUN useradd --system --no-create-home --uid 65532 nonroot

WORKDIR /app

COPY cmd/accrual/accrual_linux_amd64 /app/accrual
RUN chmod +x /app/accrual

USER nonroot

EXPOSE 8080

ENTRYPOINT ["/app/accrual"]
