FROM iron/base
ADD loadtest /
ENTRYPOINT ["./loadtest"]
CMD ["-baseAddr", "servicea", "-port", "6767", "-zuul=false", "-users", "10", "-delay", "1000"]