@echo off
::%1 repo domain hkccr.ccs.tencentyun.com
::%2 username 200027558228
::%3 password

docker login %1 --username=%2 -p %3
