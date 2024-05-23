package data

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bunotel"
	"runtime"

	"testing"
)

type Keys struct {
	Host string
	Port int32
	Username string
	Password string
	Dbname   string
	SslMode string
	TimeZone string
}

func TestNewDB(t *testing.T) {
	keys := Keys{
		Host:     "192.168.2.102",
		Port: 5432,
		Username: "dbuser_dba",
		Password: "DBUser.DBA",
		Dbname:   "users",
		SslMode :"disable",
		TimeZone :"Asia/Shanghai",
	}
	// dsn := fmt.Sprintf("postgres://dbuser_dba:DBUser.DBA@192.168.2.102:5432/users?sslmode=disable")
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s&TimeZone=%s",
		keys.Username,
		keys.Password,
		keys.Host,
		keys.Port,
		keys.Dbname,
		keys.SslMode,
		keys.TimeZone,
	)
	sqlDb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	db := bun.NewDB(sqlDb, pgdialect.New(), bun.WithDiscardUnknownColumns())

	// otel 可观测性套件
	db.AddQueryHook(bunotel.NewQueryHook(bunotel.WithDBName("meta")))

	// Go 程序的并发执行使用的 CPU 核心数设置, 0为最大可用值，并且保持当前的设置
	db.SetMaxOpenConns(4 * runtime.GOMAXPROCS(0))
	db.SetMaxIdleConns(4 * runtime.GOMAXPROCS(0))

	type User struct {
		bun.BaseModel `bun:"table:user"`

		ID int64 `bun:",pk,autoincrement"`
		Username string `bun:"username"`
		Password string `bun:"password"`
		Age      int    `bun:"age"`
		Gender   string `bun:"gender"`
	}

	if _, err := db.NewCreateTable().
		Model((*User)(nil)).
		IfNotExists().
		Exec(context.TODO());
	err != nil {
		t.Error(err)
	}
	t.Log("Success")
}

// 参考 https://redis.uptrace.dev/guide/go-redis-sentinel.html#redis-server-client
func TestNewCache(t *testing.T) {
	// 默认方式, 稳定:
	// client := redis.NewFailoverClient(&redis.FailoverOptions{
	// 	MasterName:    "master",
	// 	SentinelAddrs: []string{"192.168.2.155:26379", "192.168.2.158:26379", "192.168.2.152:26379"},
	// 	Password:      "263393", // 如果有密码，请填写
	// 	DB:            0,
	// })

	// 从 v8 开始，您可以使用实验性 NewFailoverClusterClient 命令将只读命令路由到从节点
	client := redis.NewClient(&redis.Options{
		// Addr:     "192.168.2.158:6379",
		Protocol: 3,
		Addr:     "192.168.2.192:6379",
		Username: "default", // redis实例的用户名, 非哨兵节点名
		Password: "msdnmm,.",  // redis实例的用户密码, 如果有密码，请填写
		DB:       0,
	})
	// client := redis.NewFailoverClusterClient(&redis.FailoverOptions{
	// 	// MasterName:              "master1",                                                                     // 主节点master名
	// 	// SentinelAddrs:           []string{"192.168.2.155:6379", "192.168.2.158:6379", "192.168.2.152:6379"}, // 哨兵节点
	// 	// ClientName:              "",
	// 	// SentinelUsername:        "master1", // 哨兵节点的账号
	// 	// SentinelPassword:        "263393",  // 哨兵节点的密码
	// 	// RouteByLatency:          true, // 将只读命令路由到从节点
	// 	// RouteRandomly:           true, // 将只读命令路由到从节点
	//
	// 	ReplicaOnly:             false,
	// 	UseDisconnectedReplicas: false,
	// 	Dialer:                  nil,
	// 	OnConnect:               nil,
	// 	Protocol:                0,
	// 	Username:                "master1", // redis实例的用户名, 非哨兵节点名
	// 	Password:                "263393",  // redis实例的用户密码, 如果有密码，请填写
	// 	DB:                      0,
	// 	MaxRetries:              0,
	// 	MinRetryBackoff:         0,
	// 	MaxRetryBackoff:         0,
	// 	DialTimeout:             0,
	// 	ReadTimeout:             0,
	// 	WriteTimeout:            0,
	// 	ContextTimeoutEnabled:   false,
	// 	PoolFIFO:                false,
	// 	PoolSize:                0,
	// 	PoolTimeout:             0,
	// 	MinIdleConns:            0,
	// 	MaxIdleConns:            0,
	// 	MaxActiveConns:          0,
	// 	ConnMaxIdleTime:         0,
	// 	ConnMaxLifetime:         0,
	// 	TLSConfig:               nil,
	// 	DisableIndentity:        false,
	// })

	// 设置一个键值对
	err := client.Set(context.Background(), "example_key", "example_value2", 0).Err()
	err = client.Set(context.Background(), "example_key2", "example_value3", 0).Err()
	err = client.Set(context.Background(), "example_key3", "example_value31", 0).Err()
	err = client.Set(context.Background(), "example_key4", "example_value34", 0).Err()
	if err != nil {
		panic(err)
	}

	// 获取键值对
	val, err := client.Get(context.Background(), "example_key").Result()
	val2, err := client.Get(context.Background(), "example_key2").Result()
	val3, err := client.Get(context.Background(), "example_key3").Result()
	val4, err := client.Get(context.Background(), "example_key4").Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("example_key", val)
	fmt.Println("example_key2", val2)
	fmt.Println("example_key3", val3)
	fmt.Println("example_key4", val4)

	// 关闭连接
	if err := client.Close(); err != nil {
		panic(err)
	}
}
