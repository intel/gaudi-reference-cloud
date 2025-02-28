# shellcheck shell=sh

# Best to land this in /etc/profile.d/zzz_intel_proxy.sh

# Intel Proxy recommendations
# https://circuit.intel.com/content/it/it-support/topic-pages/infrastructure-and-information-security/proxy.html
# More technical Intel Proxy info
# https://internal-placeholder.com/display/proxy/Proxy+Users+Guide+Home

# Proxy Nonsense
# https://about.gitlab.com/blog/2021/01/27/we-need-to-talk-no-proxy/

no_proxy_append() {
    if [ -n "$1" ] && [[ ",$no_proxy," != *",$1,"* ]]; then
        no_proxy="${no_proxy:+"$no_proxy,"}$1"
        NO_PROXY=${no_proxy}
#        echo "updated"
#    else
#        echo "Did not update with '$1'"
    fi
}
no_proxy_prepend() {
    if [ -n "$1" ] && [[ ",$no_proxy," != *",$1,"* ]]; then
        no_proxy="$1${no_proxy:+",$no_proxy"}"
        NO_PROXY=${no_proxy}
#        echo "updated"
#    else
#        echo "Did not update with '$1'"
    fi
}


# Are we connected to Intel?
# Check if it.intel.com can be resolved
# Surely someday IT will enable IT ticket access
# without VPN, and then this will break.
nslookup it.intel.com. 1>/dev/null 2>/dev/null
if [ "$?" -eq 0 ]; then
    #echo "Connected to Intel, setting proxies"

    # Use lowercase form. HTTP_PROXY is not always supported or recommended.

    # no_proxy
    # If the extra no_proxy values are set, they remain 
    # Use lowercase form.
    # Use comma-separated hostname:port values.
    # IP addresses are okay, but hostnames are never resolved.
    # Suffixes are always matched (e.g. example.com will match test.example.com).
    # If top-level domains need to be matched, avoid using a leading dot (.).
    # Avoid using CIDR matching since only Go and Ruby support that.
    no_proxy_prepend "intel.com"
    no_proxy_prepend ".intel.com"
    no_proxy_prepend "192.168.0.0/16"
    no_proxy_prepend "172.16.0.0/12"
    no_proxy_prepend "10.0.0.0/8"
    no_proxy_prepend "127.0.0.1"
    no_proxy_prepend "localhost"

    #_ADD_HERE_
    export no_proxy
    export ftp_proxy=ftp://internal-placeholder.com:912
    export https_proxy=http://internal-placeholder.com:912
    export http_proxy="${https_proxy}"  # Port 911 was EOL'd
    export socks_proxy=socks://internal-placeholder.com:1080
    # If you absolutely must use the uppercase form as well, be sure they share the same value.
    export NO_PROXY=${no_proxy}
    export FTP_PROXY=${ftp_proxy}
    export HTTPS_PROXY=${https_proxy}
    export HTTP_PROXY=${http_proxy}
    export SOCKS_PROXY=${socks_proxy}
else
    #echo "Not connected to Intel, removing proxies"
    typeset +x -- no_proxy ftp_proxy http_proxy https_proxy socks_proxy
    unset no_proxy ftp_proxy http_proxy https_proxy socks_proxy
    typeset +x -- NO_PROXY FTP_PROXY HTTPS_PROXY HTTP_PROXY SOCKS_PROXY
    unset NO_PROXY FTP_PROXY HTTPS_PROXY HTTP_PROXY SOCKS_PROXY
fi
