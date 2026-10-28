@echo off

CALL ".\shared\dockerCheck.bat"
if %errorlevel% neq 0 (
    echo Docker check failed. Exiting...
	timeout /t 30
    exit /b 1
)

set pWD=%CD%
set scriptRoot=%~dp0

CD %scriptRoot%
::forget all variable
CALL ".\shared\sharedEnvClean.bat"

CALL ".\sv\sharedEnv.bat"

set tcrRepo=duckrails-api-golang

:: Login with target repo
CALL ".\shared\login.bat" %tencentBase% %tcrUsername% %tcrPass%

CD ..\..\..

Set "gitver=" & For /F Delims^= %%G In ('git rev-parse --short HEAD') Do If Not Defined Git Set "gitver=%%~G"

set tag=%tencentBase%/%tcrNamespace%/%tcrRepo%

docker images --format "{{.ID}}" %tag%

set tag=%tag%:api-%gitver%

:: tag image in local repo
:: docker tag [imageId] uswccr.ccs.tencentyun.com/gr-tcr/duckrails-api-golang:[tag]
docker build --build-arg GIT_COMMIT=%gitver% --platform linux/amd64 -t %tag% -f ./Dockerfile .

:: Push image to Registry
docker push %tag%

::forget all variable
CALL "%scriptRoot%shared\sharedEnvClean.bat"

CD %pWD%

::docker image rm %tag%

timeout /t 30
@echo on