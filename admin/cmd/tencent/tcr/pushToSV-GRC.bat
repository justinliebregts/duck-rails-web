::@echo off

set serviceRegion=sv

::forget all variable
CALL "../../../../api/gofiber/cmd/tencent/tcr/shared/sharedEnvClean.bat"

CALL "../../../../api/gofiber/cmd/tencent/tcr/%serviceRegion%/sharedEnv.bat"

set tcrRepo=duckrails-api-admin

set "gitver=" & For /F Delims^= %%G In ('git rev-parse --short HEAD') Do If Not Defined Git Set "gitver=%%~G"

set tag=%tencentBase%/%tcrNamespace%/%tcrRepo%

docker images --format "{{.ID}}" %tag%

set tag=%tag%:adm-duckrails-%gitver%

:: Login with target repo
CALL "../../../../api/gofiber/cmd/tencent/tcr/shared/login.bat" %tencentBase% %tcrUsername% %tcrPass%

set preCD=%CD%

CD ../../..
:: tag image in local repo
:: docker tag [imageId] uswccr.ccs.tencentyun.com/gr-tcr/duckrails-api-admin:[tag]
docker build -t %tag% --no-cache --pull -f ./Dockerfile_sv_grc .

:: Push image to Registry
docker push %tag%

CD %preCD%
::forget all variable
CALL "../../../../api/gofiber/cmd/tencent/tcr/shared/sharedEnvClean.bat"

timeout 30
@echo on
