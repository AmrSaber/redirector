FROM alpine

HEALTHCHECK --interval=5s --start-period=30s --start-interval=1s CMD /redirector ping -q

COPY redirector /redirector
RUN chmod +x /redirector

ENTRYPOINT ["/redirector"]
