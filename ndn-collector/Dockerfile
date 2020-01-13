FROM graphiteapp/graphite-statsd AS graphite


FROM grafana/grafana:6.5.2
USER root
WORKDIR /etc/service/grafana
RUN apk add --no-cache \
        bash \
        cairo \
        collectd \
        collectd-disk \
        collectd-nginx \
        findutils \
        librrd \
        logrotate \
        memcached \
        nginx \
        nodejs \
        npm \
        py3-pyldap \
        redis \
        runit \
        sqlite \
        expect \
        dcron \
        py-mysqldb \
        mysql-dev \
        mysql-client \
        postgresql-dev \
        postgresql-client && \
    rm -rf \
        /etc/nginx/conf.d/default.conf && \
    mkdir -p \
        /var/log/carbon \
        /var/log/graphite \
        /etc/service/grafana && \
    mv /run.sh /etc/service/grafana/run
COPY --from=graphite /etc /etc
COPY --from=graphite /opt /opt
COPY --from=graphite /usr/local/bin /usr/local/bin
COPY --from=graphite /entrypoint /entrypoint
ENV STATSD_INTERFACE udp

ENTRYPOINT /entrypoint
