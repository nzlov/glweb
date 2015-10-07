package glweb

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"

	log "github.com/nzlov/glog"
)

type DB struct {
	db *sql.DB
}

func (this *DB) Open(db_user, db_pwd, db_host, db_database string) {
	//	db_host := this.yoghourt.cm.sc.Item("db_host").Value()
	//	db_user := this.yoghourt.cm.sc.Item("db_user").Value()
	//	db_pwd := this.yoghourt.cm.sc.Item("db_pwd").Value()
	//	db_database := this.yoghourt.cm.sc.Item("db_database").Value()

	if db_host != "" && db_user != "" && db_pwd != "" && db_database != "" {
		var err error
		this.db, err = sql.Open("mysql", db_user+":"+db_pwd+"@tcp("+db_host+")/"+db_database+"?charset=utf8")
		if err != nil {
			log.Errorln("DB Open", err)
		}
	} else {
		log.Errorln("DB Open", "no database info")
	}
}
func (this *DB) Close() {
	if this.db != nil {
		err := this.db.Close()
		if err != nil {
			log.Errorln("DB Close", err)
		}
	}
}

func (this *DB) DB() *sql.DB {
	return this.db
}

//插入
func (this *DB) Insert(sqlstr string, args ...interface{}) (int64, error) {
	stmtIns, err := this.db.Prepare(sqlstr)
	if err != nil {
		return 0, err
	}
	defer stmtIns.Close()

	result, err := stmtIns.Exec(args...)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

//修改和删除
func (this *DB) Exec(sqlstr string, args ...interface{}) (int64, error) {
	stmtIns, err := this.db.Prepare(sqlstr)

	if err != nil {
		return -1, err
	}
	defer stmtIns.Close()

	result, err := stmtIns.Exec(args...)
	if err != nil {
		return -2, err
	}
	return result.RowsAffected()
}

//取一行数据，注意这类取出来的结果都是string
func (this *DB) FetchRow(sqlstr string, args ...interface{}) (map[string]string, error) {
	stmtOut, err := this.db.Prepare(sqlstr)
	if err != nil {
		return nil, err
	}
	defer stmtOut.Close()

	rows, err := stmtOut.Query(args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(values))
	ret := make(map[string]string, len(scanArgs))

	for i := range values {
		scanArgs[i] = &values[i]
	}
	for rows.Next() {
		err = rows.Scan(scanArgs...)
		if err != nil {
			return nil, err
		}
		var value string

		for i, col := range values {
			if col == nil {
				value = "NULL"
			} else {
				value = string(col)
			}
			ret[columns[i]] = value
		}
		break //get the first row only
	}
	return ret, nil
}

//取多行，<span style="font-family: Arial, Helvetica, sans-serif;">注意这类取出来的结果都是string </span>
func (this *DB) FetchRows(sqlstr string, args ...interface{}) ([]map[string]string, error) {
	stmtOut, err := this.db.Prepare(sqlstr)
	if err != nil {
		return nil, err
	}
	defer stmtOut.Close()

	rows, err := stmtOut.Query(args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(values))

	ret := make([]map[string]string, 0)
	for i := range values {
		scanArgs[i] = &values[i]
	}

	for rows.Next() {
		err = rows.Scan(scanArgs...)
		if err != nil {
			return nil, err
		}
		var value string
		vmap := make(map[string]string, len(scanArgs))
		for i, col := range values {
			if col == nil {
				value = "NULL"
			} else {
				value = string(col)
			}
			vmap[columns[i]] = value
		}
		ret = append(ret, vmap)
	}
	return ret, nil
}
