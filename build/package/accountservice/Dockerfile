FROM iron/base

EXPOSE 6767
EXPOSE 8181

ADD accountservice /
ADD healthchecker /
HEALTHCHECK --interval=3s --timeout=3s CMD ["./healthchecker"] || exit 1
ENTRYPOINT ["./accountservice"]
