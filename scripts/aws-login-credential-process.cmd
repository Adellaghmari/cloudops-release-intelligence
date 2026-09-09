@echo off
REM Bridge `aws login` sessions into Terraform. Unsets AWS_CONFIG_FILE so
REM this process reads the user's real ~/.aws/config login_session.
set AWS_CONFIG_FILE=
set AWS_SDK_LOAD_CONFIG=
set AWS_PROFILE=
"C:\Program Files\Amazon\AWSCLIV2\aws.exe" configure export-credentials --format process
