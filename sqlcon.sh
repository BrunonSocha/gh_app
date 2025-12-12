docker start greyhouse-sql 2> /dev/null
docker exec -it greyhouse-sql /opt/mssql-tools18/bin/sqlcmd -S localhost -U sa -P "DavidBowie11%" -C  2> /dev/null
