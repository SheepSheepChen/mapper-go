FROM alpine
RUN mkdir -p /etc/kubeedge/exec
RUN mkdir -p /root/.kube
ADD config.yaml /etc/kubeedge/
ADD modbus-go /bin/
CMD ["/bin/modbus-go","--v","5"]