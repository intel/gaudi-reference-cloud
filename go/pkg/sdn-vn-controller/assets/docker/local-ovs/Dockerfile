FROM debian:sid

ENV http_proxy="http://proxy-dmz.intel.com:912"
ENV https_proxy="http://proxy-dmz.intel.com:912"

RUN apt-get update && \
    apt-get install -y autoconf libtool python3 gdb \
    iproute2 iputils-ping net-tools util-linux procps openvswitch-switch ovn-host && \
    apt-get clean

ENV PATH="${PATH}:/usr/share/openvswitch/scripts:/usr/share/ovn/scripts"
COPY vswitch.ovsschema /ovn/ovs/vswitchd/vswitch.ovsschema
COPY entrypoint.sh /mnt/entrypoint.sh
RUN chmod +x /mnt/entrypoint.sh

WORKDIR /mnt

ENTRYPOINT ["/mnt/entrypoint.sh"]
