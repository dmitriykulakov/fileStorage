FROM postgres:latest
COPY ./config.env ./
CMD ["postgres"]