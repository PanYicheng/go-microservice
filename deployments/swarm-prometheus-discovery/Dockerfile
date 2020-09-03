FROM iron/base

ADD swarm-prometheus-discovery /
ENTRYPOINT ["./swarm-prometheus-discovery","-network", "my_network", "-ignoredServices", "grafana,prometheus"]
