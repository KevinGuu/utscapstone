FROM ubuntu:20.04

RUN apt update -y && apt install -y iptables python3-pip
RUN pip3 install python-iptables 

ADD sidecar.py .
RUN chmod u+x sidecar.py
ENTRYPOINT ["./sidecar.py"]
