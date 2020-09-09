
package model

import (
	. "database/sql"
	. "github.com/hide2/go-sharding/db"
	. "github.com/hide2/go-sharding/lib"
	"strings"
	"time"

	"fmt"
)

type EventModel struct {
	OdB        string
	Lmt        int
	Ofs        int
	
	Datasource string
	Table      string
	AutoID     string
	Trx        *Tx
	ID         int64

	Uid int64
	Event string
	CreatedAt time.Time
}

func (m *EventModel) Begin(ds string) (*Tx, error) {
	db := DBPool[ds]["w"]
	sql := "BEGIN"
	if GoShardingSqlLog {
		fmt.Println("["+time.Now().Format("2006-01-02 15:04:05")+"][SQL]", sql)
	}
	tx, err := db.Begin()
	m.Trx = tx
	return tx, err
}

func (m *EventModel) Commit() error {
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

func (m *EventModel) Rollback() error {
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

func (m *EventModel) Exec(sql string) error {
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

func (m *EventModel) CreateTable() error {
	for i := 0; i < GoShardingDatasourceNumber; i++ {
		db := DBPool[fmt.Sprintf("ds_%d", i)]["w"]
		for j := 0; j < GoShardingTableNumber; j++ {
			table := fmt.Sprintf("event_%d", j)
			sql := fmt.Sprintf(`CREATE TABLE %s (
		id BIGINT AUTO_INCREMENT,

		uid BIGINT,
		event VARCHAR(255),
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

func (m *EventModel) New() *EventModel {
	n := EventModel{Datasource: "default", Table: "event"}
	return &n
}

func (m *EventModel) Find(id int64) (*EventModel, error) {
	db := DBPool[m.Datasource]["r"]
	sql := "SELECT * FROM event WHERE id = ?"
	if GoShardingSqlLog {
		fmt.Println("["+time.Now().Format("2006-01-02 15:04:05")+"][SQL]", sql, id)
	}
	st := time.Now().UnixNano() / 1e6
	row := db.QueryRow(sql, id)
	if err := row.Scan(&m.ID, &m.Uid, &m.Event, &m.CreatedAt); err != nil {
		return nil, err
	}
	e := time.Now().UnixNano()/1e6 - st
	if GoShardingSlowSqlLog > 0 && int(e) >= GoShardingSlowSqlLog {
		fmt.Printf("["+time.Now().Format("2006-01-02 15:04:05")+"][SlowSQL][%s][%dms]\n", sql, e)
	}
	return m, nil
}

func (m *EventModel) Save() (*EventModel, error) {
	db := DBPool[m.Datasource]["w"]
	// Update
	if m.ID > 0 {
		props := StructToMap(*m)
		conds := map[string]interface{}{"id": m.ID}
		uprops := make(map[string]interface{})
		for k, v := range props {
			if k != "OdB" && k != "Lmt" && k != "Ofs" && k != "Datasource" && k != "Table" && k != "Trx" && k != "ID" {
				uprops[Underscore(k)] = v
			}
		}
		return m, m.Update(uprops, conds)
	// Create
	} else {
		sql := "INSERT INTO event(uid,event,created_at) VALUES(?,?,?)"
		if GoShardingSqlLog {
			fmt.Println("["+time.Now().Format("2006-01-02 15:04:05")+"][SQL]", sql, m.Uid, m.Event, m.CreatedAt)
		}
		st := time.Now().UnixNano() / 1e6
		result, err := db.Exec(sql, m.Uid, m.Event, m.CreatedAt)
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

func (m *EventModel) Where(conds map[string]interface{}) ([]*EventModel, error) {
	db := DBPool[m.Datasource]["r"]
	wherestr := make([]string, 0)
	cvs := make([]interface{}, 0)
	for k, v := range conds {
		wherestr = append(wherestr, k + "=?")
		cvs = append(cvs, v)
	}
	sql := fmt.Sprintf("SELECT * FROM event WHERE %s", strings.Join(wherestr, " AND "))
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
	ms := make([]*EventModel, 0)
	for rows.Next() {
		m = new(EventModel)
		err = rows.Scan(&m.ID, &m.Uid, &m.Event, &m.CreatedAt) //不scan会导致连接不释放
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

func (m *EventModel) Create(props map[string]interface{}) (*EventModel, error) {
	if m.AutoID != "" {
		props[m.AutoID] = GenUUID()
	}
	// todo 根据sharding_column选择datasource
	db := DBPool[m.Datasource]["w"]
	if m.AutoID != "" {
		props[m.AutoID] = GenUUID()
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
	sql := fmt.Sprintf("INSERT INTO event(%s) VALUES(%s)", cstr, ph)

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
	return m.Find(lastInsertID)
}

func (m *EventModel) Delete() error {
	return m.Destroy(m.ID)
}

func (m *EventModel) Destroy(id int64) error {
	db := DBPool[m.Datasource]["w"]
	sql := "DELETE FROM event WHERE id = ?"
	if GoShardingSqlLog {
		fmt.Println("["+time.Now().Format("2006-01-02 15:04:05")+"][SQL]", sql, id)
	}
	st := time.Now().UnixNano() / 1e6
	var err error
	if m.Trx != nil {
		_, err = m.Trx.Exec(sql, id)
	} else {
		_, err = db.Exec(sql, id)
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

func (m *EventModel) Update(props map[string]interface{}, conds map[string]interface{}) error {
	if _, ok := conds[GoShardingColumn]; ok {
		// todo 根据sharding_column选择datasource
		// db := DBPool[m.Datasource]["w"]
	}
	db := DBPool[m.Datasource]["w"]
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
	sql := fmt.Sprintf("UPDATE event SET %s WHERE %s", strings.Join(setstr, ", "), strings.Join(wherestr, " AND "))
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

func (m *EventModel) CountAll() (int, error) {
	db := DBPool[m.Datasource]["r"]
	sql := "SELECT count(1) FROM event"
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

func (m *EventModel) Count(conds map[string]interface{}) (int, error) {
	db := DBPool[m.Datasource]["r"]
	wherestr := make([]string, 0)
	cvs := make([]interface{}, 0)
	for k, v := range conds {
		wherestr = append(wherestr, k + "=?")
		cvs = append(cvs, v)
	}
	sql := fmt.Sprintf("SELECT count(1) FROM event WHERE %s", strings.Join(wherestr, " AND "))
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

func (m *EventModel) All() ([]*EventModel, error) {
	db := DBPool[m.Datasource]["r"]
	sql := "SELECT * FROM event"
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
	ms := make([]*EventModel, 0)
	for rows.Next() {
		m = new(EventModel)
		err = rows.Scan(&m.ID, &m.Uid, &m.Event, &m.CreatedAt) //不scan会导致连接不释放
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

func (m *EventModel) OrderBy(o string) *EventModel {
	m.OdB = o
	return m
}

func (m *EventModel) Offset(o int) *EventModel {
	m.Ofs = o
	return m
}

func (m *EventModel) Limit(l int) *EventModel {
	m.Lmt = l
	return m
}

func (m *EventModel) Page(page int, size int) *EventModel {
	m.Ofs = (page - 1)*size
	m.Lmt = size
	return m
}

var Event = EventModel{Datasource: "default", Table: "event", AutoID: ""}
