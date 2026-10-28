package mysql

// ConnectDb connects with the local DSN; used by DB-backed tests.
func ConnectDb() {
	connectDbWithDSN(DSN)
	DefineDbBoundary(DBConn)
}
