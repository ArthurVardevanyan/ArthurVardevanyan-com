#FROM php:7-apache
FROM docker.io/httpd
#COPY . /var/www/html
COPY . /usr/local/apache2/htdocs/
EXPOSE 80
