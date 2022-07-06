FROM alpine
COPY redirector /redirector
ENTRYPOINT ["/redirector"]
