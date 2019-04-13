FROM alpine
COPY ./build/dbgp_linux_amd64 /srv/
CMD ["/srv/dbgp_linux_amd64"]