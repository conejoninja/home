FROM debian

RUN mkdir -p /data/db
RUN chmod 777 /data/db
ADD ./logger /data/logger
RUN chmod +x /data/logger
EXPOSE 8055

ENTRYPOINT cd /data/ && /data/logger