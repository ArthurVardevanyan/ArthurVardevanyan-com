#FROM php:7-apache
FROM httpd
#COPY . /var/www/html
COPY . /usr/local/apache2/htdocs/
EXPOSE 80

# docker image build -t arthurvardevanyan /home/arthur/Projects/Code/Web\ Development/ArthurVardevanyan/ -t 10.0.0.7:5000/arthurvardevanyan:latest 
# docker image tag arthurvardevanyan 10.0.0.7:5000/arthurvardevanyan:202006162033
# docker image push 10.0.0.7:5000/arthurvardevanyan:202006162033
# docker image push 10.0.0.7:5000/arthurvardevanyan:latest