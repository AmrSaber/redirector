FROM alpine
COPY redirector /redirector
RUN chmod +x /redirector
ENTRYPOINT ["/redirector"]
