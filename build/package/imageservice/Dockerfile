FROM iron/base

EXPOSE 7777

ADD imageservice /
ADD testimages/*.jpg testimages/
ADD healthchecker /

HEALTHCHECK --interval=1s --timeout=3s CMD ["./healthchecker"] || exit 1

ENTRYPOINT ["./imageservice"]
