@echo off
setlocal
cd /d "%~dp0"

set serviceRegion=sv

::forget all variable
CALL "../../../../api/gofiber/cmd/tencent/tcr/shared/sharedEnvClean.bat"

CALL "../../../../api/gofiber/cmd/tencent/tcr/%serviceRegion%/sharedEnv.bat"
@echo off

set tcrRepo=duckrails-api-admin

set "gitver=" & For /F Delims^= %%G In ('git rev-parse --short HEAD') Do If Not Defined Git Set "gitver=%%~G"

set tag=%tencentBase%/%tcrNamespace%/%tcrRepo%

docker images --format "{{.ID}}" %tag%

set tag=%tag%:adm-duckrails-%gitver%

:: Login with target repo
CALL "../../../../api/gofiber/cmd/tencent/tcr/shared/login.bat" %tencentBase% %tcrUsername% %tcrPass%
if errorlevel 1 exit /b 1

set preCD=%CD%

CD ../../..
:: tag image in local repo
:: docker tag [imageId] uswccr.ccs.tencentyun.com/gr-tcr/duckrails-api-admin:[tag]
docker build -t %tag% --no-cache --pull -f ./Dockerfile_sv_grc .
if errorlevel 1 exit /b 1

:: Push image to Registry
docker push %tag%
if errorlevel 1 exit /b 1

CD %preCD%
::forget all variable
CALL "../../../../api/gofiber/cmd/tencent/tcr/shared/sharedEnvClean.bat"

endlocal
exit /b 0
