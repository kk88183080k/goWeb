package orm

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/kk88183080k/goWeb/msgo/logs"
	"log"
	"reflect"
	"strings"
	"time"
)

// MsDb : 相等于对数据库连接池进行了一层包装
type MsDb struct {
	db     *sql.DB
	prefix string
	log    *logs.Logger
}

type MsDbSession struct {
	db          *MsDb
	TableName   string           // 表名
	FieldName   []string         // 字段名
	placeHolder []string         // 占位符
	values      []any            // 字段值
	where       []any            // where条件字段值
	updateSql   *strings.Builder // 字段值拼接的update set sql
	whereSql    *strings.Builder // where条件字段值拼接的where条件
	tx          *sql.Tx          // 事务
	beginTx     bool             // 是否开启事务
}

/*--------------数据库(连接池)开始----------------------*/

func Open(driverName string, source string) *MsDb {
	db, err := sql.Open(driverName, source)
	if err != nil {
		panic(err)
	}
	err = db.Ping()
	if err != nil {
		panic(err)
	}

	// 连接最大存活时间
	db.SetConnMaxLifetime(time.Minute * 3)
	// 连接最大空间时间
	db.SetConnMaxIdleTime(time.Minute * 1)
	// 最大转接数
	db.SetMaxOpenConns(100)
	// 最大空闲连接数
	db.SetMaxIdleConns(10)

	msDb := &MsDb{
		db:  db,
		log: logs.Default(),
	}

	return msDb
}

func (msdb *MsDb) SetMaxIdleConntions(n int) {
	msdb.db.SetMaxIdleConns(n)
}

func (msdb *MsDb) SetTablePrefix(tablePrefix string) {
	msdb.prefix = tablePrefix
}

func (msdb *MsDb) Close() {
	msdb.db.Close()
}

func (msdb *MsDb) New(data any) *MsDbSession {
	dataType := reflect.TypeOf(data)
	if dataType.Kind() != reflect.Pointer {
		panic(errors.New("data type is not Pointer "))
	}

	// 结构体类名中大写的，要用下划线分隔
	session := msdb.NewSessionByTableName(msdb.prefix + strings.ToLower(Name(dataType.Elem().Name())))
	return session
}

func (msdb *MsDb) NewSessionByTableName(tableName string) *MsDbSession {
	session := &MsDbSession{db: msdb, TableName: tableName, updateSql: &strings.Builder{}, whereSql: &strings.Builder{}}
	return session
}

/*--------------数据库(连接池)结束----------------------*/

/*--------------数据库session 开始----------------------*/

func (session *MsDbSession) Table(tableName string) *MsDbSession {
	session.TableName = tableName
	return session
}

// Insert insert 返回: id, 影响的行数，错误
func (session *MsDbSession) Insert(v any) (int64, int64, error) {
	// 使用反射生成表的字段，占位符列表，
	session.fieldName(v)
	sql := fmt.Sprintf("insert into %s(%s) values(%s)", session.TableName, strings.Join(session.FieldName, ","), strings.Join(session.placeHolder, ","))
	return session.execute(sql, session.values...)
}

func (session *MsDbSession) InsertBatch(v []any) (int64, int64, error) {
	if v == nil || len(v) == 0 {
		return -1, -1, errors.New("批量添加参能不能为空")
	}

	// insert into 表名(列名...) values (),()
	session.fieldName(v[0])
	sql := fmt.Sprintf("insert into %s(%s) values ", session.TableName, strings.Join(session.FieldName, ","))
	sqlBuilder := &strings.Builder{}
	fmt.Fprintf(sqlBuilder, sql)

	// 生成values部分
	var values []any
	for i, val := range v {
		if i != 0 {
			fmt.Fprintf(sqlBuilder, ",")
		}
		fmt.Fprintf(sqlBuilder, "(")
		fmt.Fprintf(sqlBuilder, strings.Join(session.placeHolder, ","))
		fmt.Fprintf(sqlBuilder, ")")

		t := reflect.TypeOf(val)
		if t.Kind() != reflect.Pointer {
			return -1, -1, errors.New("InsertBatch data is not Pointer ")
		}
		vType := t.Elem()
		value := reflect.ValueOf(val).Elem()
		for i := 0; i < vType.NumField(); i++ {
			// 首字母是小写的不做处理
			if !value.Field(i).CanInterface() {
				continue
			}

			// 根据tag进行处理
			field := vType.Field(i)
			sqlTag := field.Tag.Get("msgo")
			// 对自增标记做处理
			contains := strings.Contains(sqlTag, "auto_increment")
			if field.Name == "Id" || contains {
				if isAutoId(value.Field(i).Interface()) {
					continue
				}
			}
			values = append(values, value.Field(i).Interface())
		}
	}

	return session.execute(sqlBuilder.String(), values...)
}

// UpdateField 添加更新字段 update table set 字段=?,字段=? where id=? and age =?
func (session *MsDbSession) UpdateField(fieldName string, v any) *MsDbSession {
	session.values = append(session.values, v)
	if session.updateSql.Len() > 0 {
		fmt.Fprintf(session.updateSql, " ,")
	}
	fmt.Fprintf(session.updateSql, fieldName+"=?")
	return session
}

func (session *MsDbSession) Update() (int64, error) {
	sql := &strings.Builder{}
	fmt.Fprintf(sql, "update %s set %s %s", session.TableName, session.updateSql.String(), session.whereSql.String())

	var val []any
	val = append(val, session.values...)
	val = append(val, session.where...)
	_, rows, err := session.execute(sql.String(), val...)
	return rows, err
}

func (session *MsDbSession) UpdateObject(data any) (int64, error) {
	//fmt.Fprintf(sql, "update %s set %s %s", session.TableName, session.updateSql.String(), session.whereSql.String())
	t := reflect.TypeOf(data)
	if t.Kind() != reflect.Pointer {
		return -1, errors.New("update data is not pointer")
	}

	tc := t.Elem()
	val := reflect.ValueOf(data).Elem()
	// 生成update setSql & paraList
	// 生成update whereSql & paraList
	for i := 0; i < tc.NumField(); i++ {
		field := tc.Field(i)

		if !val.Field(i).CanInterface() {
			continue
		}

		sqlTag := field.Tag.Get("msgo")
		if sqlTag == "" {
			sqlTag = field.Name
		}

		contains := strings.Contains(sqlTag, "auto_increment")
		if contains {
			// 根据主键更新
			sqlTag = sqlTag[:strings.Index(sqlTag, ",")]
		}

		if contains || field.Name == "Id" {
			session.And().Where(strings.ToLower(Name(sqlTag)), val.Field(i).Interface())
			continue
		}
		session.UpdateField(strings.ToLower(Name(sqlTag)), val.Field(i).Interface())
	}

	return session.Update()
}

func (session *MsDbSession) DeleteObject(data any) (int64, error) {
	// 生成where条件
	t := reflect.TypeOf(data)
	if t.Kind() != reflect.Pointer {
		return -1, errors.New("delete data type is not pointer")
	}

	te := t.Elem()
	v := reflect.ValueOf(data).Elem()
	for i := 0; i < te.NumField(); i++ {
		if !v.Field(i).CanInterface() {
			continue
		}

		field := te.Field(i)
		sqlTag := field.Tag.Get("msgo")
		contains := strings.Contains(sqlTag, "auto_increment")
		if contains || field.Name == "Id" { // 是主键
			if sqlTag == "" {
				sqlTag = field.Name
			}
			if contains {
				sqlTag = sqlTag[:strings.Index(sqlTag, ",")]
			}
			session.Where(sqlTag, v.Field(i).Interface())
		}
	}

	return session.Delete()
}

func (session *MsDbSession) Delete() (int64, error) {
	if session.whereSql.Len() <= 0 {
		return -1, errors.New("where condition is empty")
	}

	sql := &strings.Builder{}
	fmt.Fprintf(sql, "delete from %s %s", session.TableName, session.whereSql)

	_, rows, err := session.execute(sql.String(), session.where...)
	return rows, err
}

func (session MsDbSession) SelectOne(entity any, filedName ...string) error {
	t := reflect.TypeOf(entity)
	if t.Kind() != reflect.Pointer {
		return errors.New("entity type is not Pointer ")
	}
	// select * from %s where
	filedList := "*"
	if len(filedName) > 0 {
		filedList = strings.Join(filedName, ",")
	}

	sql := &strings.Builder{}
	fmt.Fprintf(sql, "select %s from %s %s limit 1", filedList, session.TableName, session.whereSql.String())
	session.db.log.Info("query sql:" + sql.String())
	stmt, err := session.db.db.Prepare(sql.String())
	if err != nil {
		return err
	}
	result, err := stmt.Query(session.where...)
	if err != nil {
		return err
	}

	columns, err := result.Columns()
	if err != nil {
		return err
	}

	values := make([]any, len(columns))
	filedScan := make([]any, len(columns))
	for i := 0; i < len(filedScan); i++ {
		filedScan[i] = &values[i]
	}
	// 是否有查询结果
	if result.Next() {
		err := result.Scan(filedScan...)
		if err != nil {
			return err
		}

		v := reflect.ValueOf(entity)
		dbValue := reflect.ValueOf(values)
		tElem := t.Elem()
		for i := 0; i < tElem.NumField(); i++ {
			field := tElem.Field(i)
			sqlTag := field.Tag.Get("msgo")
			if sqlTag == "" {
				sqlTag = strings.ToLower(Name(field.Name))
			}
			contains := strings.Contains(sqlTag, ",")
			if contains {
				sqlTag = sqlTag[:strings.Index(sqlTag, ",")]
			}

			for j := 0; j < len(columns); j++ {
				if sqlTag == columns[j] {
					if v.Elem().Field(i).CanSet() {
						covertValue := session.ConvertType(dbValue, j, v, i)
						v.Elem().Field(i).Set(covertValue)
					}
				}
			}
		}
	}
	return nil
}

func (session MsDbSession) SelectList(entity any, filedName ...string) ([]any, error) {
	t := reflect.TypeOf(entity)
	if t.Kind() != reflect.Pointer {
		return nil, errors.New("entity type is not Pointer ")
	}
	// select * from %s where
	filedList := "*"
	if len(filedName) > 0 {
		filedList = strings.Join(filedName, ",")
	}

	sql := &strings.Builder{}
	fmt.Fprintf(sql, "select %s from %s %s", filedList, session.TableName, session.whereSql.String())
	session.db.log.Info("query sql:" + sql.String())
	stmt, err := session.db.db.Prepare(sql.String())
	if err != nil {
		return nil, err
	}
	result, err := stmt.Query(session.where...)
	if err != nil {
		return nil, err
	}

	columns, err := result.Columns()
	if err != nil {
		return nil, err
	}

	values := make([]any, len(columns))
	filedScan := make([]any, len(columns))
	for i := 0; i < len(filedScan); i++ {
		filedScan[i] = &values[i]
	}

	var results []any
	for {
		// 是否有查询结果
		if result.Next() {
			err := result.Scan(filedScan...)
			if err != nil {
				return nil, err
			}

			// 通过反射new一个对象出来
			entity = reflect.New(t.Elem()).Interface()
			v := reflect.ValueOf(entity)
			dbValue := reflect.ValueOf(values)
			tElem := t.Elem()
			for i := 0; i < tElem.NumField(); i++ {
				field := tElem.Field(i)
				sqlTag := field.Tag.Get("msgo")
				if sqlTag == "" {
					sqlTag = strings.ToLower(Name(field.Name))
				}
				contains := strings.Contains(sqlTag, ",")
				if contains {
					sqlTag = sqlTag[:strings.Index(sqlTag, ",")]
				}

				for j := 0; j < len(columns); j++ {
					if sqlTag == columns[j] {
						if v.Elem().Field(i).CanSet() {
							covertValue := session.ConvertType(dbValue, j, v, i)
							v.Elem().Field(i).Set(covertValue)
						}
					}
				}
			}

			results = append(results, entity)
		} else {
			// 处理完之后，跳出循环
			break
		}
	}

	return results, nil
}

func (s *MsDbSession) ConvertType(valueOf reflect.Value, j int, v reflect.Value, i int) reflect.Value {
	eVar := valueOf.Index(j)
	t2 := v.Elem().Field(i).Type()
	of := reflect.ValueOf(eVar.Interface())
	covertValue := of.Convert(t2)
	return covertValue
}

func (session *MsDbSession) Where(fieldName string, v any) *MsDbSession {
	if len(session.whereSql.String()) <= 0 {
		fmt.Fprintf(session.whereSql, " where ")
	}
	fmt.Fprintf(session.whereSql, fieldName+" =?")

	session.where = append(session.where, v)
	return session
}

func (session *MsDbSession) And() *MsDbSession {
	if session.whereSql.Len() > 0 {
		fmt.Fprintf(session.whereSql, " and ")
	}
	return session
}

func (session *MsDbSession) Or() *MsDbSession {
	if session.whereSql.Len() > 0 {
		fmt.Fprintf(session.whereSql, " or ")
	}

	return session
}

func (session *MsDbSession) execute(sql string, val ...any) (int64, int64, error) {
	log.Println("sql:", sql)
	stmt, err := session.db.db.Prepare(sql)
	if err != nil {
		return -1, -1, err
	}
	result, err := stmt.Exec(val...)
	if err != nil {
		return -1, -1, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return -1, -1, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return -1, -1, err
	}
	return id, affected, nil
}

func (session *MsDbSession) Begin() error {
	tx, err := session.db.db.Begin()
	if err != nil {
		return err
	}

	session.tx = tx
	session.beginTx = true

	return nil
}

func (session *MsDbSession) Commit() error {
	err := session.tx.Commit()
	if err != nil {
		return err
	}

	session.beginTx = false
	return nil
}

func (session *MsDbSession) Rollback() error {
	err := session.tx.Rollback()
	if err != nil {
		return err
	}

	session.beginTx = false
	return nil
}

/*--------------数据库session 结束----------------------*/

/*--------------数据库session 实体 开始----------------------*/
func (s *MsDbSession) fieldName(data any) {
	dataType := reflect.TypeOf(data)
	dataVal := reflect.ValueOf(data)

	if dataType.Kind() != reflect.Pointer {
		panic(errors.New("data must is pointer"))
	}

	dataTypeClass := dataType.Elem()
	dataValInstance := dataVal.Elem()
	if s.TableName == "" {
		s.TableName = s.db.prefix + strings.ToLower(Name(dataTypeClass.Name()))
	}

	var fieldNames []string   // 字段名
	var placeHolders []string // 占位符
	var values []any          // 字段值
	for i := 0; i < dataTypeClass.NumField(); i++ {
		// 首字母是小写的不做处理
		if !dataValInstance.Field(i).CanInterface() {
			continue
		}

		// 根据tag进行处理
		var fieldName string
		field := dataTypeClass.Field(i)
		sqlTag := field.Tag.Get("msgo")
		if sqlTag == "" {
			fieldName = strings.ToLower(Name(field.Name))
		}
		// 对自增标记做处理
		contains := strings.Contains(sqlTag, "auto_increment")
		if field.Name == "Id" || contains {
			if isAutoId(dataValInstance.Field(i).Interface()) {
				continue
			}
		}
		if contains {
			fieldName = sqlTag[:strings.Index(sqlTag, ",")]
		}

		if fieldName == "" {
			fieldName = sqlTag
		}

		fieldNames = append(fieldNames, fieldName)
		placeHolders = append(placeHolders, "?")
		values = append(values, dataValInstance.Field(i).Interface())
	}
	// 设置值
	s.FieldName = fieldNames
	s.placeHolder = placeHolders
	s.values = values
}

func isAutoId(id any) bool {
	t := reflect.TypeOf(id)
	v := reflect.ValueOf(id)
	switch t.Kind() {
	case reflect.Int64:
		if v.Interface().(int64) <= 0 {
			return true
		}
	case reflect.Int32:
		if v.Interface().(int32) <= 0 {
			return true
		}
	case reflect.Int16:
		if v.Interface().(int16) <= 0 {
			return true
		}
	default:
		return false
	}
	return false
}

func Name(name string) string {
	all := name[:]
	builder := &strings.Builder{}

	for i, v := range all {
		if v >= 65 && v <= 90 { // 这些是大写的英文字符
			if i != 0 { // 第一个字符不处理
				fmt.Fprintf(builder, "_")
			}
		}
		fmt.Fprintf(builder, all[i:i+1])
	}

	return builder.String()
}

/*--------------数据库session 实体 结束----------------------*/
