FROM golang:1.27

WORKDIR /usr/src/app

ENV PATH="/usr/local/go/bin:${PATH}"
COPY . .
RUN go build -o url-shortner .

EXPOSE 8080
CMD [ "./url-shortner" ]