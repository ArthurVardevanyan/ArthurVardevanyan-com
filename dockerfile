#FROM php:7-apache
FROM httpd
#COPY . /var/www/html
COPY . /usr/local/apache2/htdocs/
EXPOSE 80
