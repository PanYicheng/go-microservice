FROM iron/base

EXPOSE 6767

ADD customservice / 
ADD healthchecker /
# HEALTHCHECK --interval=3s --timeout=3s CMD ["./healthchecker"] || exit 1
ENTRYPOINT ["./customservice"]
