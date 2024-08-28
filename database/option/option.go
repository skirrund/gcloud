package option

type Option struct {
	DSN string
	//打开数据库连接的最大数量，默认DefaultMaxOpenConns
	MaxOpenConns int
	//空闲连接池中连接的最大数量，默认DefaultMaxIdleConns
	MaxIdleConns int
	//连接可复用的最大时间。单位分钟，默认DefaultConnMaxLifetime
	ConnMaxLifetime int
	//QueryFields executes the SQL query with all fields of the table
	QueryFields bool
	//数据源类型：MySql
	Type string
}
