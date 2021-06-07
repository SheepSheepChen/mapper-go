FROM alpine
RUN mkdir -p /etc/kubeedge/exec
ADD config.yaml /etc/kubeedge/
ADD modbus-go /bin/
CMD ["/bin/modbus-go","--v","5"]