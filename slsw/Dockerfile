FROM golang AS lwe
COPY . /usr/local/src
WORKDIR /usr/local/src
RUN make build-slsw

FROM alpine
COPY --from=lwe /usr/local/src/bin/simuate-single-website /usr/local/bin/
CMD /usr/local/bin/simuate-single-website
