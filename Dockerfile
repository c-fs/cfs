FROM scratch

ADD cfs /
ADD server/default.conf /

EXPOSE 15524
ENTRYPOINT ["/cfs"]
