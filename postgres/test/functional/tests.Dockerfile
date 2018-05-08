FROM scratch

COPY ./bin/functional.test /

CMD ["/functional.test", "-test.v"]