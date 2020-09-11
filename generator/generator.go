package generator

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"text/template"

	. "github.com/hide2/go-sharding/db"
	. "github.com/hide2/go-sharding/lib"

	"gopkg.in/yaml.v2"
)

const GOORMTEMPLATE = `
package model

import (
	. "database/sql"
	. "github.com/hide2/go-sharding/db"
	. "github.com/hide2/go-sharding/lib"
	"strings"
	"errors"
	"time"
{{ range $i, $m := .Imports }}
	"{{$m}}"
{{- end }}
)

type {{.Model}}Model struct {
	OdB        string
	Lmt        int
	Ofs        int
	
	Datasource string
	Table      string
	AutoID     string
	Trx        *Tx
	ID         int64
{{ range $i, $k := .Attrs }}
	{{$k}} {{index $.Values $i}}
{{- end }}
}

func (m *{{.Model}}Model) Begin(ds string) error {
	if ds == "" {
		return errors.New("Error Begin, need ds string")
	}
	db := DBPool[ds]["w"]
	sql := "BEGIN"
	if GoShardingSqlLog {
		fmt.Println("["+time.Now().Format("2006-01-02 15:04:05")+"][SQL]", sql)
	}
	tx, err := db.Begin()
	m.Trx = tx
	return err
}

func (m *{{.Model}}Model) Commit() error {
	if m.Trx != nil {
		sql := "COMMIT"
		if GoShardingSqlLog {
			fmt.Println("["+time.Now().Format("2006-01-02 15:04:05")+"][SQL]", sql)
		}
		return m.Trx.Commit()
	}
	m.Trx = nil
	return nil
}

func (m *{{.Model}}Model) Rollback() error {
	if m.Trx != nil {
		sql := "ROLLBACK"
		if GoShardingSqlLog {
			fmt.Println("["+time.Now().Format("2006-01-02 15:04:05")+"][SQL]", sql)
		}
		return m.Trx.Rollback()
	}
	m.Trx = nil
	return nil
}

func (m *{{.Model}}Model) Exec(sql string) error {
	for i := 0; i < GoShardingDatasourceNumber; i++ {
		db := DBPool[fmt.Sprintf("ds_%d", i)]["w"]
		if GoShardingSqlLog {
			fmt.Println("["+time.Now().Format("2006-01-02 15:04:05")+"][SQL]", sql)
		}
		st := time.Now().UnixNano() / 1e6
		if _, err := db.Exec(sql); err != nil {
			fmt.Println("Execute sql failed:", err)
			return err
		}
		e := time.Now().UnixNano()/1e6 - st
		if GoShardingSlowSqlLog > 0 && int(e) >= GoShardingSlowSqlLog {
			fmt.Printf("["+time.Now().Format("2006-01-02 15:04:05")+"][SlowSQL][%s][%dms]\n", sql, e)
		}
	}
	return nil
}

func (m *{{.Model}}Model) CreateTable() error {
	for i := 0; i < GoShardingDatasourceNumber; i++ {
		db := DBPool[fmt.Sprintf("ds_%d", i)]["w"]
		for j := 0; j < GoShardingTableNumber; j++ {
			table := fmt.Sprintf("{{.Table}}_%d", j)
			sql := ` + "fmt.Sprintf(`" + `CREATE TABLE %s (
		id BIGINT AUTO_INCREMENT,
{{ range $i, $k := .Keys }}
		{{$k}} {{index $.Columns $i}},
{{- end }}
		PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;` + "`, table)" + `
			if GoShardingSqlLog {
				fmt.Println("["+time.Now().Format("2006-01-02 15:04:05")+"][SQL]", sql)
			}
			st := time.Now().UnixNano() / 1e6
			if _, err := db.Exec(sql); err != nil {
				fmt.Println("Create table failed:", err)
				return err
			}
			e := time.Now().UnixNano()/1e6 - st
			if GoShardingSlowSqlLog > 0 && int(e) >= GoShardingSlowSqlLog {
				fmt.Printf("["+time.Now().Format("2006-01-02 15:04:05")+"][SlowSQL][%s][%dms]\n", sql, e)
			}
		}
	}
	return nil
}

func (m *{{.Model}}Model) New() *{{.Model}}Model {
	n := {{.Model}}Model{Datasource: "default", Table: "{{.Table}}"}
	return &n
}

{{if .ShardingID}}
func (m *{{.Model}}Model) FindBy{{.ShardingID}}(sid int64) (*{{.Model}}Model, error) {
	ds_fix := sid / int64(GoShardingTableNumber) % int64(GoShardingDatasourceNumber)
	table_fix := sid % int64(GoShardingTableNumber)
	ds := fmt.Sprintf("ds_%d", ds_fix)
	table := fmt.Sprintf("{{.Table}}_%d", table_fix)
	m.Datasource = ds
	m.Table = table
	sharding_column := Underscore("{{.ShardingID}}")

	db := DBPool[ds]["r"]
	sql := fmt.Sprintf("SELECT * FROM %s WHERE %s = ?", table, sharding_column)

	if GoShardingSqlLog {
		fmt.Println("["+time.Now().Format("2006-01-02 15:04:05")+"][SQL]", sql, sid)
	}
	st := time.Now().UnixNano() / 1e6
	row := db.QueryRow(sql, sid)
	if err := row.Scan({{.ScanStr}}); err != nil {
		return nil, err
	}
	e := time.Now().UnixNano()/1e6 - st
	if GoShardingSlowSqlLog > 0 && int(e) >= GoShardingSlowSqlLog {
		fmt.Printf("["+time.Now().Format("2006-01-02 15:04:05")+"][SlowSQL][%s][%dms]\n", sql, e)
	}
	return m, nil
}
{{end}}

{{if .ShardingID}}
func (m *{{.Model}}Model) FindBy{{.ShardingID}}AndID(sid int64, id int64) (*{{.Model}}Model, error) {
	ds_fix := sid / int64(GoShardingTableNumber) % int64(GoShardingDatasourceNumber)
	table_fix := sid % int64(GoShardingTableNumber)
	ds := fmt.Sprintf("ds_%d", ds_fix)
	table := fmt.Sprintf("{{.Table}}_%d", table_fix)
	m.Datasource = ds
	m.Table = table
	sharding_column := Underscore("{{.ShardingID}}")

	db := DBPool[ds]["r"]
	sql := fmt.Sprintf("SELECT * FROM %s WHERE %s = ? and id = ?", table, sharding_column)

	if GoShardingSqlLog {
		fmt.Println("["+time.Now().Format("2006-01-02 15:04:05")+"][SQL]", sql, sid, id)
	}
	st := time.Now().UnixNano() / 1e6
	row := db.QueryRow(sql, sid, id)
	if err := row.Scan({{.ScanStr}}); err != nil {
		return nil, err
	}
	e := time.Now().UnixNano()/1e6 - st
	if GoShardingSlowSqlLog > 0 && int(e) >= GoShardingSlowSqlLog {
		fmt.Printf("["+time.Now().Format("2006-01-02 15:04:05")+"][SlowSQL][%s][%dms]\n", sql, e)
	}
	return m, nil
}
{{end}}

func (m *{{.Model}}Model) Save() (*{{.Model}}Model, error) {
	// Update
	if m.ID > 0 {
		props := StructToMap(*m)
		conds := map[string]interface{}{"id": m.ID, Underscore("{{.ShardingID}}"): m.{{.ShardingID}}}
		uprops := make(map[string]interface{})
		for k, v := range props {
			if k != "OdB" && k != "Lmt" && k != "Ofs" && k != "Datasource" && k != "Table" && k != "Trx" && k != "ID" && k != "AutoID" && k != "{{.ShardingID}}" {
				uprops[Underscore(k)] = v
			}
		}
		return m, m.Update(uprops, conds)
	// Create
	} else {
		ds_fix := int64(0)
		table := "{{.Table}}"

		{{if .AutoID}}
		// Gen UUID
		sid := GenUUID()
		m.{{.AutoID}} = sid
		ds_fix = sid / int64(GoShardingTableNumber) % int64(GoShardingDatasourceNumber)
		table_fix := sid % int64(GoShardingTableNumber)
		ds := fmt.Sprintf("ds_%d", ds_fix)
		table = fmt.Sprintf("{{.Table}}_%d", table_fix)
		m.Datasource = ds
		m.Table = table

		{{else if .ShardingID}}

		// ShardingColumn
		sid := m.{{.ShardingID}}
		if sid <= 0 {
			return nil, errors.New("Error Save, no Sharding Column")
		}
		ds_fix = sid / int64(GoShardingTableNumber) % int64(GoShardingDatasourceNumber)
		table_fix := sid % int64(GoShardingTableNumber)
		ds := fmt.Sprintf("ds_%d", ds_fix)
		table = fmt.Sprintf("{{.Table}}_%d", table_fix)
		m.Datasource = ds
		m.Table = table

		{{else}}
		return nil, errors.New("Error Save, no Sharding Column")
		{{end}}

		sql := "{{.InsertSQL}}"
		sql = strings.Replace(sql, "table", table, 1)

		db := DBPool[ds]["w"]

		if GoShardingSqlLog {
			fmt.Println("["+time.Now().Format("2006-01-02 15:04:05")+"][SQL]", sql, {{.InsertArgs}})
		}
		st := time.Now().UnixNano() / 1e6
		result, err := db.Exec(sql, {{.InsertArgs}})
		if err != nil {
			fmt.Printf("Insert data failed, err:%v\n", err)
			return nil, err
		}
		lastInsertID, err := result.LastInsertId() //获取插入数据的自增ID
		if err != nil {
			fmt.Printf("Get insert id failed, err:%v\n", err)
			return nil, err
		}
		m.ID = lastInsertID
		e := time.Now().UnixNano()/1e6 - st
		if GoShardingSlowSqlLog > 0 && int(e) >= GoShardingSlowSqlLog {
			fmt.Printf("["+time.Now().Format("2006-01-02 15:04:05")+"][SlowSQL][%s][%dms]\n", sql, e)
		}
	}
	return m, nil
}

func (m *{{.Model}}Model) Where(conds map[string]interface{}) ([]*{{.Model}}Model, error) {
	var ds, table string
	if sid, ok := conds[GoShardingColumn]; ok {
		ds_fix := sid.(int64)  / int64(GoShardingTableNumber) % int64(GoShardingDatasourceNumber)
		table_fix := sid.(int64) % int64(GoShardingTableNumber)
		ds = fmt.Sprintf("ds_%d", ds_fix)
		table = fmt.Sprintf("{{.Table}}_%d", table_fix)
	
		m.Datasource = ds
		m.Table = table
		db := DBPool[ds]["r"]

		wherestr := make([]string, 0)
		cvs := make([]interface{}, 0)
		for k, v := range conds {
			wherestr = append(wherestr, k + "=?")
			cvs = append(cvs, v)
		}
		sql := fmt.Sprintf("SELECT * FROM %s WHERE %s", table, strings.Join(wherestr, " AND "))
		if m.OdB != "" {
			sql = sql + " ORDER BY " + m.OdB
		}
		if m.Lmt > 0 {
			sql = sql + fmt.Sprintf(" LIMIT %d", m.Lmt)
		}
		if m.Ofs > 0 {
			sql = sql + fmt.Sprintf(" OFFSET %d", m.Ofs)
		}
		// Clear Limitation
		m.OdB = ""
		m.Lmt = 0
		m.Ofs = 0
		if GoShardingSqlLog {
			fmt.Println("["+time.Now().Format("2006-01-02 15:04:05")+"][SQL]", sql, cvs)
		}
		st := time.Now().UnixNano() / 1e6
		rows, err := db.Query(sql, cvs...)
		defer func() {
			if rows != nil {
				rows.Close() //关闭掉未scan的sql连接
			}
		}()
		if err != nil {
			fmt.Printf("Query data failed, err:%v\n", err)
			return nil, err
		}
		ms := make([]*{{.Model}}Model, 0)
		for rows.Next() {
			m = new({{.Model}}Model)
			err = rows.Scan({{.ScanStr}}) //不scan会导致连接不释放
			if err != nil {
				return nil, err
			}
			m.Datasource = ds
			m.Table = table
			ms = append(ms, m)
		}
		e := time.Now().UnixNano()/1e6 - st
		if GoShardingSlowSqlLog > 0 && int(e) >= GoShardingSlowSqlLog {
			fmt.Printf("["+time.Now().Format("2006-01-02 15:04:05")+"][SlowSQL][%s][%dms]\n", sql, e)
		}
		return ms, nil
	} else {
		// No Sharding Column,  Find Across All Databases & Tables
		ms := make([]*{{.Model}}Model, 0)
		for i := 0; i < GoShardingDatasourceNumber; i++ {
			db := DBPool[fmt.Sprintf("ds_%d", i)]["r"]
			for j := 0; j < GoShardingTableNumber; j++ {
				table = fmt.Sprintf("{{.Table}}_%d", j)

				wherestr := make([]string, 0)
				cvs := make([]interface{}, 0)
				for k, v := range conds {
					wherestr = append(wherestr, k + "=?")
					cvs = append(cvs, v)
				}
				sql := fmt.Sprintf("SELECT * FROM %s WHERE %s", table, strings.Join(wherestr, " AND "))
				// if m.OdB != "" {
				// 	sql = sql + " ORDER BY " + m.OdB
				// }
				// if m.Lmt > 0 {
				// 	sql = sql + fmt.Sprintf(" LIMIT %d", m.Lmt)
				// }
				// if m.Ofs > 0 {
				// 	sql = sql + fmt.Sprintf(" OFFSET %d", m.Ofs)
				// }
				if GoShardingSqlLog {
					fmt.Println("["+time.Now().Format("2006-01-02 15:04:05")+"][SQL]", sql, cvs)
				}
				st := time.Now().UnixNano() / 1e6
				rows, err := db.Query(sql, cvs...)
				defer func() {
					if rows != nil {
						rows.Close() //关闭掉未scan的sql连接
					}
				}()
				if err != nil {
					fmt.Printf("Query data failed, err:%v\n", err)
					return nil, err
				}
				
				for rows.Next() {
					m = new({{.Model}}Model)
					err = rows.Scan({{.ScanStr}}) //不scan会导致连接不释放
					if err != nil {
						return nil, err
					}
					m.Datasource = ds
					m.Table = table
					ms = append(ms, m)
				}
				e := time.Now().UnixNano()/1e6 - st
				if GoShardingSlowSqlLog > 0 && int(e) >= GoShardingSlowSqlLog {
					fmt.Printf("["+time.Now().Format("2006-01-02 15:04:05")+"][SlowSQL][%s][%dms]\n", sql, e)
				}
			}
		}
		return ms, nil
	}
}

func (m *{{.Model}}Model) Create(props map[string]interface{}) (*{{.Model}}Model, error) {
	ds_fix := int64(0)
	table := "{{.Table}}"

	{{if .AutoID}}
	// Gen UUID
	sid :=  GenUUID()
	props[Underscore(m.AutoID)] = sid
	m.{{.AutoID}} = sid
	ds_fix = sid / int64(GoShardingTableNumber) % int64(GoShardingDatasourceNumber)
	table_fix := sid % int64(GoShardingTableNumber)
	ds := fmt.Sprintf("ds_%d", ds_fix)
	table = fmt.Sprintf("{{.Table}}_%d", table_fix)
	m.Datasource = ds
	m.Table = table

	{{else if .ShardingID}}

	// ShardingColumn
	sid := props[GoShardingColumn].(int64)
	if sid <= 0 {
		return nil, errors.New("Error Create, no Sharding Column")
	}
	m.{{.ShardingID}} = sid
	ds_fix = sid / int64(GoShardingTableNumber) % int64(GoShardingDatasourceNumber)
	table_fix := sid % int64(GoShardingTableNumber)
	ds := fmt.Sprintf("ds_%d", ds_fix)
	table = fmt.Sprintf("{{.Table}}_%d", table_fix)
	m.Datasource = ds
	m.Table = table

	{{else}}
	return nil, errors.New("Error Create, no Sharding Column")
	{{end}}

	db := DBPool[ds]["w"]

	keys := make([]string, 0)
	values := make([]interface{}, 0)
	for k, v := range props {
		keys = append(keys, k)
		values = append(values, v)
	}
	cstr := strings.Join(keys, ",")
	phs := make([]string, 0)
	for i := 0; i < len(keys); i++ {
		phs = append(phs, "?")
	}
	ph := strings.Join(phs, ",")
	sql := fmt.Sprintf("INSERT INTO %s(%s) VALUES(%s)", table, cstr, ph)

	if GoShardingSqlLog {
		fmt.Println("["+time.Now().Format("2006-01-02 15:04:05")+"][SQL]", sql, values)
	}
	st := time.Now().UnixNano() / 1e6
	var result Result
	var err error
	if m.Trx != nil {
		result, err = m.Trx.Exec(sql, values...)
	} else {
		result, err = db.Exec(sql, values...)
	}
	if err != nil {
		fmt.Printf("Insert data failed, err:%v\n", err)
		return nil, err
	}
	lastInsertID, err := result.LastInsertId() //获取插入数据的自增ID
	if err != nil {
		fmt.Printf("Get insert id failed, err:%v\n", err)
		return nil, err
	}
	e := time.Now().UnixNano()/1e6 - st
	if GoShardingSlowSqlLog > 0 && int(e) >= GoShardingSlowSqlLog {
		fmt.Printf("["+time.Now().Format("2006-01-02 15:04:05")+"][SlowSQL][%s][%dms]\n", sql, e)
	}
	
	return m.FindBy{{.ShardingID}}AndID(sid, lastInsertID)
}

func (m *{{.Model}}Model) Delete() error {
	return m.DestroyBy{{.ShardingID}}(m.{{.ShardingID}})
}

func (m *{{.Model}}Model) DestroyBy{{.ShardingID}}(sid int64) error {
	{{if .ShardingID}}
	ds_fix := sid / int64(GoShardingTableNumber) % int64(GoShardingDatasourceNumber)
	table_fix := sid % int64(GoShardingTableNumber)
	ds := fmt.Sprintf("ds_%d", ds_fix)
	table := fmt.Sprintf("{{.Table}}_%d", table_fix)
	m.Datasource = ds
	m.Table = table
	sharding_column := Underscore("{{.ShardingID}}")

	db := DBPool[ds]["w"]
	sql := fmt.Sprintf("DELETE FROM %s WHERE %s = ?", table, sharding_column)
	if GoShardingSqlLog {
		fmt.Println("["+time.Now().Format("2006-01-02 15:04:05")+"][SQL]", sql, sid)
	}
	st := time.Now().UnixNano() / 1e6
	var err error
	if m.Trx != nil {
		_, err = m.Trx.Exec(sql, sid)
	} else {
		_, err = db.Exec(sql, sid)
	}
	if err != nil {
		fmt.Printf("Delete data failed, err:%v\n", err)
		return err
	}
	m.ID = 0
	e := time.Now().UnixNano()/1e6 - st
	if GoShardingSlowSqlLog > 0 && int(e) >= GoShardingSlowSqlLog {
		fmt.Printf("["+time.Now().Format("2006-01-02 15:04:05")+"][SlowSQL][%s][%dms]\n", sql, e)
	}
	return nil
	{{else}}
	return errors.New("Error Destroy, no Sharding Column")
	{{end}}
}

func (m *{{.Model}}Model) Update(props map[string]interface{}, conds map[string]interface{}) error {
	var ds, table string
	if sid, ok := conds[GoShardingColumn]; ok {
		ds_fix := sid.(int64)  / int64(GoShardingTableNumber) % int64(GoShardingDatasourceNumber)
		table_fix := sid.(int64) % int64(GoShardingTableNumber)
		ds = fmt.Sprintf("ds_%d", ds_fix)
		table = fmt.Sprintf("{{.Table}}_%d", table_fix)
	} else {
		return errors.New("Error Update, no Sharding Column")
	}
	m.Datasource = ds
	m.Table = table
	db := DBPool[ds]["w"]
	setstr := make([]string, 0)
	wherestr := make([]string, 0)
	cvs := make([]interface{}, 0)
	for k, v := range props {
		setstr = append(setstr, k + "=?")
		cvs = append(cvs, v)
	}
	for k, v := range conds {
		wherestr = append(wherestr, k + "=?")
		cvs = append(cvs, v)
	}
	sql := fmt.Sprintf("UPDATE %s SET %s WHERE %s", table, strings.Join(setstr, ", "), strings.Join(wherestr, " AND "))
	if GoShardingSqlLog {
		fmt.Println("["+time.Now().Format("2006-01-02 15:04:05")+"][SQL]", sql, cvs)
	}
	st := time.Now().UnixNano() / 1e6
	var err error
	if m.Trx != nil {
		_, err = m.Trx.Exec(sql, cvs...)
	} else {
		_, err = db.Exec(sql, cvs...)
	}
	if err != nil {
		fmt.Printf("Update data failed, err:%v\n", err)
		return err
	}
	e := time.Now().UnixNano()/1e6 - st
	if GoShardingSlowSqlLog > 0 && int(e) >= GoShardingSlowSqlLog {
		fmt.Printf("["+time.Now().Format("2006-01-02 15:04:05")+"][SlowSQL][%s][%dms]\n", sql, e)
	}
	return nil
}

func (m *{{.Model}}Model) CountAll() (int, error) {
	// No Sharding Column,  Count Across All Databases & Tables
	cc := 0
	for i := 0; i < GoShardingDatasourceNumber; i++ {
		db := DBPool[fmt.Sprintf("ds_%d", i)]["r"]
		for j := 0; j < GoShardingTableNumber; j++ {
			table := fmt.Sprintf("{{.Table}}_%d", j)
			sql := fmt.Sprintf("SELECT count(1) FROM %s", table)
			if GoShardingSqlLog {
				fmt.Println("["+time.Now().Format("2006-01-02 15:04:05")+"][SQL]", sql)
			}
			st := time.Now().UnixNano() / 1e6
			row := db.QueryRow(sql)
			var c int
			if err := row.Scan(&c); err != nil {
				return 0, err
			}
			e := time.Now().UnixNano()/1e6 - st
			if GoShardingSlowSqlLog > 0 && int(e) >= GoShardingSlowSqlLog {
				fmt.Printf("["+time.Now().Format("2006-01-02 15:04:05")+"][SlowSQL][%s][%dms]\n", sql, e)
			}
			cc = cc + c
		}
	}
	return cc, nil
}

func (m *{{.Model}}Model) Count(conds map[string]interface{}) (int, error) {
	cc := 0
	var ds, table string
	if sid, ok := conds[GoShardingColumn]; ok {
		ds_fix := sid.(int64)  / int64(GoShardingTableNumber) % int64(GoShardingDatasourceNumber)
		table_fix := sid.(int64) % int64(GoShardingTableNumber)
		ds = fmt.Sprintf("ds_%d", ds_fix)
		table = fmt.Sprintf("{{.Table}}_%d", table_fix)
	
		m.Datasource = ds
		m.Table = table
		db := DBPool[ds]["r"]

		wherestr := make([]string, 0)
		cvs := make([]interface{}, 0)
		for k, v := range conds {
			wherestr = append(wherestr, k + "=?")
			cvs = append(cvs, v)
		}
		sql := fmt.Sprintf("SELECT count(1) FROM %s WHERE %s", table, strings.Join(wherestr, " AND "))
		if GoShardingSqlLog {
			fmt.Println("["+time.Now().Format("2006-01-02 15:04:05")+"][SQL]", sql, cvs)
		}
		st := time.Now().UnixNano() / 1e6
		row := db.QueryRow(sql, cvs...)
		var c int
		if err := row.Scan(&c); err != nil {
			return 0, err
		}
		e := time.Now().UnixNano()/1e6 - st
		if GoShardingSlowSqlLog > 0 && int(e) >= GoShardingSlowSqlLog {
			fmt.Printf("["+time.Now().Format("2006-01-02 15:04:05")+"][SlowSQL][%s][%dms]\n", sql, e)
		}
		cc = c
	} else {
		// No Sharding Column,  Count Across All Databases & Tables
		for i := 0; i < GoShardingDatasourceNumber; i++ {
			db := DBPool[fmt.Sprintf("ds_%d", i)]["r"]
			for j := 0; j < GoShardingTableNumber; j++ {
				table = fmt.Sprintf("{{.Table}}_%d", j)
				wherestr := make([]string, 0)
				cvs := make([]interface{}, 0)
				for k, v := range conds {
					wherestr = append(wherestr, k + "=?")
					cvs = append(cvs, v)
				}
				sql := fmt.Sprintf("SELECT count(1) FROM %s WHERE %s", table, strings.Join(wherestr, " AND "))
				if GoShardingSqlLog {
					fmt.Println("["+time.Now().Format("2006-01-02 15:04:05")+"][SQL]", sql, cvs)
				}
				st := time.Now().UnixNano() / 1e6
				row := db.QueryRow(sql, cvs...)
				var c int
				if err := row.Scan(&c); err != nil {
					return 0, err
				}
				e := time.Now().UnixNano()/1e6 - st
				if GoShardingSlowSqlLog > 0 && int(e) >= GoShardingSlowSqlLog {
					fmt.Printf("["+time.Now().Format("2006-01-02 15:04:05")+"][SlowSQL][%s][%dms]\n", sql, e)
				}
				cc = cc + c
			}
		}
	}
	return cc, nil
}

func (m *{{.Model}}Model) OrderBy(o string) *{{.Model}}Model {
	m.OdB = o
	return m
}

func (m *{{.Model}}Model) Offset(o int) *{{.Model}}Model {
	m.Ofs = o
	return m
}

func (m *{{.Model}}Model) Limit(l int) *{{.Model}}Model {
	m.Lmt = l
	return m
}

func (m *{{.Model}}Model) Page(page int, size int) *{{.Model}}Model {
	m.Ofs = (page - 1)*size
	m.Lmt = size
	return m
}

var {{.Model}} = {{.Model}}Model{Datasource: "default", Table: "{{.Table}}", AutoID: "{{.AutoID}}"}
`

var inputConfigFile = flag.String("file", "model.yml", "Input model config yaml file")

type ModelAttr struct {
	Model      string
	Table      string
	Imports    []string
	Attrs      []string
	Keys       []string
	Values     []string
	Columns    []string
	InsertSQL  string
	InsertArgs string
	ScanStr    string
	AutoID     string
	ShardingID string
}

func Gen() {
	flag.Parse()
	for i := 0; i != flag.NArg(); i++ {
		fmt.Printf("arg[%d]=%s\n", i, flag.Arg(i))
	}

	mf, _ := ioutil.ReadFile(*inputConfigFile)
	ms := make(map[string][]yaml.MapSlice)
	merr := yaml.Unmarshal(mf, &ms)
	if merr != nil {
		fmt.Println("error:", merr)
	}
	for _, j := range ms["models"] {
		var modelname, table, filename string
		imports := make([]string, 0)
		attrs := make([]string, 0)
		keys := make([]string, 0)
		values := make([]string, 0)
		columns := make([]string, 0)
		imports = append(imports, "fmt")
		autoid := ""
		shardingid := ""
		for _, v := range j {
			if v.Key != "model" {
				attrs = append(attrs, Camelize(v.Key.(string)))
				keys = append(keys, v.Key.(string))
				if v.Key.(string) == GoShardingColumn {
					shardingid = Camelize(v.Key.(string))
				}
				c := v.Value.(string)
				if c == "int64|auto" {
					values = append(values, "int64")
				} else {
					values = append(values, c)
				}
				if c == "string" {
					c = "VARCHAR(255)"
				} else if c == "int64" {
					c = "BIGINT"
				} else if c == "int64|auto" {
					c = "BIGINT"
					autoid = Camelize(v.Key.(string))
				} else if c == "time.Time" {
					c = "DATETIME"
					// imports = append(imports, "time")
				}
				columns = append(columns, c)
			} else {
				modelname = v.Value.(string)
				table = strings.ToLower(modelname)
				filename = "model/" + modelname + ".go"
			}
		}
		fmt.Println("-- Generate", filename)
		t, err := template.New("GOORMTEMPLATE").Parse(GOORMTEMPLATE)
		if err != nil {
			fmt.Println(err)
			return
		}
		cstr := strings.Join(keys, ",")
		phs := make([]string, 0)
		iargs := make([]string, 0)
		scans := make([]string, 0)
		scans = append(scans, "&m.ID")
		for i := 0; i < len(attrs); i++ {
			phs = append(phs, "?")
			iargs = append(iargs, "m."+attrs[i])
			scans = append(scans, "&m."+attrs[i])
		}
		ph := strings.Join(phs, ",")
		iarg := strings.Join(iargs, ", ")
		scanstr := strings.Join(scans, ", ")
		isql := fmt.Sprintf("INSERT INTO table(%s) VALUES(%s)", cstr, ph)
		m := ModelAttr{modelname, table, imports, attrs, keys, values, columns, isql, iarg, scanstr, autoid, shardingid}
		var b bytes.Buffer
		t.Execute(&b, m)
		// fmt.Println(b.String())

		// Write to file
		f, err := os.Create(filename)
		if err != nil {
			fmt.Println("create file: ", err)
			return
		}
		err = t.Execute(f, m)
		if err != nil {
			fmt.Print("execute: ", err)
			return
		}
		f.Close()

	}

}
