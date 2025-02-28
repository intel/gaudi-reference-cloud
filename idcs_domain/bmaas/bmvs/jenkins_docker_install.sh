#!/bin/bash
source /etc/os-release

if [ "$EUID" -ne 0 ]; then
    echo "Please use sudo or run as root"
    exit 1
fi

if [ "$NAME" == "Ubuntu" ]; then

    # Works around for tzdata install hang
    export TZ=UTC
    ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

    touch /etc/apt/apt.conf.d/99fixbadproxy
    echo "Acquire::http::Pipeline-Depth 0;" >> /etc/apt/apt.conf.d/99fixbadproxy
    echo "Acquire::http::No-Cache true;" >> /etc/apt/apt.conf.d/99fixbadproxy
    echo "Acquire::BrokenProxy true;" >> /etc/apt/apt.conf.d/99fixbadproxy
    echo "# [intel-linux] Proxy Bypass for Linux Security Updates" >> /etc/apt/apt.conf.d/99fixbadproxy
    echo "Acquire::http::Proxy::ftp.osuosl.org "DIRECT";" >> /etc/apt/apt.conf.d/99fixbadproxy
    echo "Acquire::http::Proxy::security.unbuntu.com "DIRECT";" >> /etc/apt/apt.conf.d/99fixbadproxy
    echo "Acquire::http::Proxy::us.archive.ubuntu.com "DIRECT";" >> /etc/apt/apt.conf.d/99fixbadproxy
    echo "Acquire::https::Proxy::ftp.osuosl.org "DIRECT";" >> /etc/apt/apt.conf.d/99fixbadproxy
    echo "Acquire::https::Proxy::security.unbuntu.com "DIRECT";" >> /etc/apt/apt.conf.d/99fixbadproxy
    echo "Acquire::https::Proxy::us.archive.ubuntu.com "DIRECT";" >> /etc/apt/apt.conf.d/99fixbadproxy
    apt-get -q update -o Acquire::CompressionTypes::Order::=gz && apt-get -q upgrade -y
    apt-get clean
    rm -rf /var/lib/apt/lists/*
    apt-get -q update -y
    apt-get -q upgrade -y
    apt-get -q install ansible-lint make git -y

    # # Install Intel root certificates
    # CERTPATH=/usr/local/share/ca-certificates
    # cd ${CERTPATH}
    # wget -q --no-proxy http://certificates.intel.com/repository/certificates/IntelSHA2RootChain-Base64.zip
    # wget -q --no-proxy http://certificates.intel.com/repository/certificates/Public%20Root%20Certificate%20Chain%20Base64.zip
    # unzip "Public Root Certificate Chain Base64.zip"
    # unzip "IntelSHA2RootChain-Base64.zip"

    # # Refresh certificates
    # update-ca-certificates --fresh
    # cd -

    # # Handle python requirements
    # python3 -m pip install --upgrade pip
    # pip3 install -q -r requirements.txt
fi
