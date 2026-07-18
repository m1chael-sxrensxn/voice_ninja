FROM ghcr.io/meetecho/janus-gateway:latest

# Create directories for Kubernetes-mounted config/logs
RUN mkdir -p /etc/janus /var/log/janus

EXPOSE 8088 8089 8188 10000-10200/udp

CMD ["/usr/local/bin/janus", "-F", "/etc/janus"]