FROM scratch
COPY redirector /redirector
ENTRYPOINT ["/redirector"]
