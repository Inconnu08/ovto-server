@ECHO OFF 

TITLE CockroachDB Server
ECHO Running CRDB...
ECHO ============================
cockroach start --insecure
ECHO ============================

PAUSE