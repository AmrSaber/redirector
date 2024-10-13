FROM alpine

HEALTHCHECK CMD /redirector ping -q

COPY redirector /redirector
RUN chmod +x /redirector

ENTRYPOINT ["/redirector"]
