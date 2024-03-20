FROM ubuntu:20.04
COPY vbook /app/vbook
WORKDIR /app
CMD ["/app/vbook"]