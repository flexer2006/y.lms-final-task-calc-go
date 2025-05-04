FROM nginx:1.27.5-alpine

RUN apk add --no-cache curl

COPY deploy/nginx/nginx.conf /etc/nginx/nginx.conf.template
COPY deploy/nginx/ssl /etc/nginx/ssl
COPY scripts/nginx-entry.sh /docker-entrypoint.sh

RUN chmod +x /docker-entrypoint.sh
RUN mkdir -p /var/www/certbot

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://${NGINX_HOST:-0.0.0.0}/health || exit 1

EXPOSE 80 443

ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["nginx", "-g", "daemon off;"]