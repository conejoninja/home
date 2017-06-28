FROM debian

RUN mkdir -p /data/db
RUN chmod 777 /data/db
ADD ./api /data/api
RUN chmod +x /data/api
EXPOSE 8080

ENTRYPOINT cd /data/ && /data/api