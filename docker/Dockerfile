FROM debian

RUN mkdir -p /data/db
RUN chmod 777 /data/db
ADD ./home /data/home
RUN chmod +x /data/home


RUN set -x \
    && add-apt-repository ppa:mc3man/trusty-media \
    && apt-get update \
    && apt-get dist-upgrade \
    && apt-get install -y --no-install-recommends \
        ffmpeg \ 

ENTRYPOINT cd /data/ && /data/home