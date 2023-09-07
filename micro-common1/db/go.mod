module common/db

go 1.18

require (
	common/log v0.0.0
	common/config v0.0.0
	github.com/go-sql-driver/mysql v1.5.0
	github.com/jmoiron/sqlx v1.2.0
)

replace common/log => ../log
replace common/config => ../config