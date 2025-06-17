@echo off
set "USER_EEC_CONFIG_DIR=D:\win\program\go\main-project\eec\eec-configs"
eec run --tag dev --program cmd --program-args="/K cd D:\win\program && D: && checkitems %USER_EEC_CONFIG_DIR%\checkitems.csv"