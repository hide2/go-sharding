
package model

import (
	. "database/sql"
	. "github.com/hide2/go-sharding/db"
	. "github.com/hide2/go-sharding/lib"
	"strings"
	"errors"
	"time"

	"fmt"
)

type UserModel struct {
	OdB        string
	Lmt        int
	Ofs        int
	
	Datasource string
	Table      string
	AutoID     string
	Trx        *Tx
	ID         int64

	Uid int64
	Name string
	CreatedAt time.Time
}

func (m *UserModel) Begin(ds string) error {
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

func (m *UserModel) Commit() error {
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

func (m *UserModel) Rollback() error {
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

func (m *UserModel) Exec(sql string) error {
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

func (m *UserModel) CreateTable() error {
	for i := 0; i < GoShardingDatasourceNumber; i++ {
		db := DBPool[fmt.Sprintf("ds_%d", i)]["w"]
		for j := 0; j < GoShardingTableNumber; j++ {
			table := fmt.Sprintf("user_%d", j)
			sql := fmt.Sprintf(`CREATE TABLE %s (
		id BIGINT AUTO_INCREMENT,

		uid BIGINT,
		name VARCHAR(255),
		created_at DATETIME,
		PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`, table)
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

func (m *UserModel) New() *UserModel {
	n := UserModel{Datasource: "default", Table: "user"}
	return &n
}


func (m *UserModel) FindByUid(sid int64) (*UserModel, error) {
	ds_fix := sid / int64(GoShardingTableNumber) % int64(GoShardingDatasourceNumber)
	table_fix := sid % int64(GoShardingTableNumber)
	ds := fmt.Sprintf("ds_%d", ds_fix)
	table := fmt.Sprintf("user_%d", table_fix)
	m.Datasource = ds
	m.Table = table
	sharding_column := Underscore("Uid")

	db := DBPool[ds]["r"]
	sql := fmt.Sprintf("SELECT * FROM %s WHERE %s = ?", table, sharding_column)

	if GoShardingSqlLog {
		fmt.Println("["+time.Now().Format("2006-01-02 15:04:05")+"][SQL]", sql, sid)
	}
	st := time.Now().UnixNano() / 1e6
	row := db.QueryRow(sql, sid)
	if err := row.Scan(&m.ID, &m.Uid, &m.Name, &m.CreatedAt); err != nil {
		return nil, err
	}
	e := time.Now().UnixNano()/1e6 - st
	if GoShardingSlowSqlLog > 0 && int(e) >= GoShardingSlowSqlLog {
		fmt.Printf("["+time.Now().Format("2006-01-02 15:04:05")+"][SlowSQL][%s][%dms]\n", sql, e)
	}
	return m, nil
}


func (m *UserModel) Save() (*UserModel, error) {
	// Update
	if m.ID > 0 {
		props := StructToMap(*m)
		conds := map[string]interface{}{"id": m.ID, Underscore("Uid"): m.Uid}
		uprops := make(map[string]interface{})
		for k, v := range props {
			if k != "OdB" && k != "Lmt" && k != "Ofs" && k != "Datasource" && k != "Table" && k != "Trx" && k != "ID" && k != "AutoID" && k != "Uid" {
				uprops[Underscore(k)] = v
			}
		}
		return m, m.Update(uprops, conds)
	// Create
	} else {
		table := "user"
		ds_fix := int64(0)

		
		// Gen UUID
		m.Uid = GenUUID()
		ds_fix = int64(m.Uid) / int64(GoShardingTableNumber) % int64(GoShardingDatasourceNumber)
		table_fix := int64(m.Uid) % int64(GoShardingTableNumber)
		ds := fmt.Sprintf("ds_%d", ds_fix)
		table = fmt.Sprintf("user_%d", table_fix)
		m.Datasource = ds
		m.Table = table

		

		sql := "INSERT INTO table(uid,name,created_at) VALUES(?,?,?)"
		sql = strings.Replace(sql, "table", table, 1)

		db := DBPool[ds]["w"]

		if GoShardingSqlLog {
			fmt.Println("["+time.Now().Format("2006-01-02 15:04:05")+"][SQL]", sql, m.Uid, m.Name, m.CreatedAt)
		}
		st := time.Now().UnixNano() / 1e6
		result, err := db.Exec(sql, m.Uid, m.Name, m.CreatedAt)
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

func (m *UserModel) Where(conds map[string]interface{}) ([]*UserModel, error) {
	// todo
	db := DBPool[m.Datasource]["r"]
	wherestr := make([]string, 0)
	cvs := make([]interface{}, 0)
	for k, v := range conds {
		wherestr = append(wherestr, k + "=?")
		cvs = append(cvs, v)
	}
	sql := fmt.Sprintf("SELECT * FROM user WHERE %s", strings.Join(wherestr, " AND "))
	if m.OdB != "" {
		sql = sql + " ORDER BY " + m.OdB
	}
	if m.Lmt > 0 {
		sql = sql + fmt.Sprintf(" LIMIT %d", m.Lmt)
	}
	if m.Ofs > 0 {
		sql = sql + fmt.Sprintf(" OFFSET %d", m.Ofs)
	}
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
	ms := make([]*UserModel, 0)
	for rows.Next() {
		m = new(UserModel)
		err = rows.Scan(&m.ID, &m.Uid, &m.Name, &m.CreatedAt) //不scan会导致连接不释放
		if err != nil {
			return nil, err
		}
		ms = append(ms, m)
	}
	e := time.Now().UnixNano()/1e6 - st
	if GoShardingSlowSqlLog > 0 && int(e) >= GoShardingSlowSqlLog {
		fmt.Printf("["+time.Now().Format("2006-01-02 15:04:05")+"][SlowSQL][%s][%dms]\n", sql, e)
	}
	return ms, nil
}

func (m *UserModel) Create(props map[string]interface{}) (*UserModel, error) {
	if m.AutoID != "" {
		props[Underscore(m.AutoID)] = GenUUID()
	}
	// todo 根据sharding_column选择datasource
	db := DBPool[m.Datasource]["w"]
	if m.AutoID != "" {
		props[Underscore(m.AutoID)] = GenUUID()
	}
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
	sql := fmt.Sprintf("INSERT INTO user(%s) VALUES(%s)", cstr, ph)

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
	m.ID = lastInsertID
	e := time.Now().UnixNano()/1e6 - st
	if GoShardingSlowSqlLog > 0 && int(e) >= GoShardingSlowSqlLog {
		fmt.Printf("["+time.Now().Format("2006-01-02 15:04:05")+"][SlowSQL][%s][%dms]\n", sql, e)
	}
	return m, nil
}

func (m *UserModel) Delete() error {
	return m.DestroyByUid(m.Uid)
}

func (m *UserModel) DestroyByUid(sid int64) error {
	
	ds_fix := sid / int64(GoShardingTableNumber) % int64(GoShardingDatasourceNumber)
	table_fix := sid % int64(GoShardingTableNumber)
	ds := fmt.Sprintf("ds_%d", ds_fix)
	table := fmt.Sprintf("user_%d", table_fix)
	m.Datasource = ds
	m.Table = table
	sharding_column := Underscore("Uid")

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
	
}

func (m *UserModel) Update(props map[string]interface{}, conds map[string]interface{}) error {
	var ds, table string
	if s, ok := conds[GoShardingColumn]; ok {
		ds_fix := s.(int64)  / int64(GoShardingTableNumber) % int64(GoShardingDatasourceNumber)
		table_fix := s.(int64) % int64(GoShardingTableNumber)
		ds = fmt.Sprintf("ds_%d", ds_fix)
		table = fmt.Sprintf("user_%d", table_fix)
	} else {
		fmt.Println("Error Update, no Sharding Column", conds)
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

func (m *UserModel) CountAll() (int, error) {
	// todo
	db := DBPool[m.Datasource]["r"]
	sql := "SELECT count(1) FROM user"
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
	return c, nil
}

func (m *UserModel) Count(conds map[string]interface{}) (int, error) {
	// todo
	db := DBPool[m.Datasource]["r"]
	wherestr := make([]string, 0)
	cvs := make([]interface{}, 0)
	for k, v := range conds {
		wherestr = append(wherestr, k + "=?")
		cvs = append(cvs, v)
	}
	sql := fmt.Sprintf("SELECT count(1) FROM user WHERE %s", strings.Join(wherestr, " AND "))
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
	return c, nil
}

func (m *UserModel) All() ([]*UserModel, error) {
	// todo
	db := DBPool[m.Datasource]["r"]
	sql := "SELECT * FROM user"
	if m.OdB != "" {
		sql = sql + " ORDER BY " + m.OdB
	}
	if m.Lmt > 0 {
		sql = sql + fmt.Sprintf(" LIMIT %d", m.Lmt)
	}
	if m.Ofs > 0 {
		sql = sql + fmt.Sprintf(" OFFSET %d", m.Ofs)
	}
	if GoShardingSqlLog {
		fmt.Println("["+time.Now().Format("2006-01-02 15:04:05")+"][SQL]", sql)
	}
	st := time.Now().UnixNano() / 1e6
	rows, err := db.Query(sql)
	defer func() {
		if rows != nil {
			rows.Close() //关闭掉未scan的sql连接
		}
	}()
	if err != nil {
		fmt.Printf("Query data failed, err:%v\n", err)
		return nil, err
	}
	ms := make([]*UserModel, 0)
	for rows.Next() {
		m = new(UserModel)
		err = rows.Scan(&m.ID, &m.Uid, &m.Name, &m.CreatedAt) //不scan会导致连接不释放
		if err != nil {
			return nil, err
		}
		ms = append(ms, m)
	}
	e := time.Now().UnixNano()/1e6 - st
	if GoShardingSlowSqlLog > 0 && int(e) >= GoShardingSlowSqlLog {
		fmt.Printf("["+time.Now().Format("2006-01-02 15:04:05")+"][SlowSQL][%s][%dms]\n", sql, e)
	}
	return ms, nil
}

var User = UserModel{Datasource: "default", Table: "user", AutoID: "Uid"}
