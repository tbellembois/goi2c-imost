FROM golang:1.16-buster
LABEL author="Thomas Bellembois"

COPY notifier /var/www-data/
COPY entrypoint.sh /

RUN chmod +x /entrypoint.sh

USER www-data
WORKDIR /var/www-data
ENTRYPOINT [ "/entrypoint.sh" ]
VOLUME ["/csv"]