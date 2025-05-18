FROM debian:bullseye-slim
# Create directories for certificates and binary
RUN mkdir -p /certs /bin
# Copy SSL certificates
COPY ./certs/nesasia.io.combined.crt /certs/
COPY ./certs/nesasia.io_key.txt /certs/
# Copy the compiled binary
COPY ./object-detection-zero-shot /bin/
# Create upload directory for images
RUN mkdir -p /uploads && chmod 755 /uploads
# Set working directory
WORKDIR /bin
# Expose HTTPS port
EXPOSE 443
# Run the binary
CMD ["./object-detection-zero-shot", "-service"]