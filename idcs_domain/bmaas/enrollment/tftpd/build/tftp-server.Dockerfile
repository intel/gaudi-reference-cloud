FROM alpine:3.19.3

# Package will be downloaded manually since armhf has no package for syslinux (#1).
ARG SYSLINUX_PACKAGE="https://dl-cdn.alpinelinux.org/alpine/v3.17/main/x86_64/syslinux-6.04_pre1-r11.apk"

RUN apk update && apk upgrade --available && \
    apk add --no-cache tftp-hpa=5.2-r7 curl=8.9.0-r0

#Setting up the basic pxelinux environment.
RUN mkdir /tmp/syslinux && \
    curl "$SYSLINUX_PACKAGE" -o /tmp/syslinux/syslinux.apk  && \
    tar -C /tmp/syslinux -xvf /tmp/syslinux/syslinux.apk && \
    mkdir -m 0755 /tftpboot && \
    cp -r /tmp/syslinux/usr/share/syslinux /tftpboot && \
    rm -rf /tmp/syslinux && \
    find /tftpboot -type f -exec chmod 444 {} \;  && \
    find /tftpboot -mindepth 1 -type d -exec chmod 555 {} \;  && \
    ln -s ../boot /tftpboot/syslinux/boot && \
    ln -s ../boot /tftpboot/syslinux/efi64/boot


COPY ipxe-snponly-x86-64.efi /tftpboot/ipxe-snponly-x86-64.efi
COPY snponly.efi /tftpboot/snponly.efi

EXPOSE 69/udp

CMD ["in.tftpd" , "-L" , "-vvv" , "-u", "ftp" ,"--secure" , "--address", "0.0.0.0:69", "/tftpboot"]
