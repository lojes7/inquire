#!/bin/sh
host_ip=127.0.0.1
port=7897

PROXY_HTTP="http://${host_ip}:${port}"
PROXY_SOCKS="socks5://${host_ip}:${port}"

set_proxy() {
    export http_proxy="${PROXY_HTTP}"
    export HTTP_PROXY="${PROXY_HTTP}"
    export https_proxy="${PROXY_HTTP}"
    export HTTPS_proxy="${PROXY_HTTP}"

    export all_proxy="$PROXY_SOCKS"
    export ALL_PROXY="$PROXY_SOCKS"

    export no_proxy="localhost,127.0.0.1,::1"
    export NO_PROXY="localhost,127.0.0.1,::1"
}

set_proxy_github() {
    git config --global http.proxy "${PROXY_HTTP}"
    git config --global https.proxy "${PROXY_SOCKS}"
}

unset_proxy() {
    unset http_proxy
    unset HTTP_PROXY
    unset https_proxy
    unset HTTPS_PROXY
    unset all_proxy
    unset ALL_PROXY
    unset no_proxy
    unset NO_PROXY
}

unset_proxy_github() {
    git config --global --unset http.proxy
    git config --global --unset https.proxy
}

test_setting(){
    echo "Host ip:" ${host_ip}
    echo "Port no: " $port
    echo "Current proxy: " $https_proxy
}


if [ "$1" = "set" ]
then
    set_proxy

elif [ "$1" = "setg" ]
then
    set_proxy_github

elif [ "$1" = "unset" ]
then
    unset_proxy

elif [ "$1" = "unsetg" ]
then
    unset_proxy_github

elif [ "$1" = "test" ]
then
    test_setting
else
    echo "Unknown arguments. Supported arguments are: set|unset|setg|unsetg|test"
fi
