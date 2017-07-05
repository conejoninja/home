FROM debian

RUN mkdir -p /data/db
RUN chmod 777 /data/db
ADD ./home /data/home
RUN chmod +x /data/home

ENTRYPOINT cd /data/ && /data/home