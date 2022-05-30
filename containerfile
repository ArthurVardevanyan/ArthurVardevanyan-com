#FROM php:7-apache
FROM docker.io/httpd
#COPY . /var/www/html

ARG quay_expiration=24h
LABEL quay.expires-after=${quay_expiration}

COPY . /usr/local/apache2/htdocs/
EXPOSE 80
