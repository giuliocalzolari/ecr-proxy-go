FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

# Run as non-root user
USER nonroot:nonroot
# Copy the binary and certificates from builder
COPY ecr-proxy .

# Environment variables with defaults
ENV AWS_REGION=us-east-1
ENV AWS_ACCOUNT_ID=
ENV PROXY_PORT=5000

# Expose HTTPS port
EXPOSE ${PROXY_PORT}


LABEL org.opencontainers.image.description="ECR Proxy for AWS ECR"
LABEL org.opencontainers.image.url="github.com/giuliocalzolari/ecr-proxy"
ENTRYPOINT ["/app/ecr-proxy"]
