package mysql

import (
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/jinzhu/gorm"
	"github.com/orvice/shadowsocks-go/mu/log"
	"github.com/orvice/shadowsocks-go/mu/user"
	ss "github.com/orvice/shadowsocks-go/shadowsocks"
)

var client *Client

func SetClient(c *Client) {
	client = c
}

func NewClient() *Client {
	mclient := new(Client)
	return mclient
}

type Client struct {
	db    *gorm.DB
	table string
}

type User struct {
	id             int
	port           int
	passwd         string
	method         string
	enable         int
	transferEnable int
	u              int
	d              int
	//结束时间
	end_time        int
	//现在时间
	now_time        int
	//套餐类型
	package_type   int
}

func (c *Client) SetDb(db *gorm.DB) {
	c.db = db
}

func (c User) TableName() string {
	return tableName
}

func (u *User) GetPort() int {
	return u.port
}

func (u *User) GetPasswd() string {
	return u.passwd
}

func (u *User) GetMethod() string {
	return u.method
}

func (u *User) IsEnable() bool {
	if u.enable == 0 {
		return false
	}
	if u.u+u.d > u.transferEnable {
		return false
	}
	//如果过期，停用用户
	if u.now_time > u.end_time {
		return false
	}
	return true
}

func (u *User) GetCipher() (*ss.Cipher, error, bool) {
	method := u.method
	auth := false

	if strings.HasSuffix(method, "-auth") {
		method = method[:len(method)-5]
		auth = true
	}
	s, e := ss.NewCipher(method, u.passwd)
	return s, e, auth
}

func (u *User) UpdateTraffic(storageSize int) error {
	return client.db.Model(u).Where("id = ?", u.id).UpdateColumn("d", gorm.Expr("d  + ?", storageSize)).UpdateColumn("t", time.Now().Unix()).Error
}

func (u *User) GetUserInfo() user.UserInfo {
	return user.UserInfo{
		Passwd: u.passwd,
		Port:   u.port,
		Method: u.method,
	}
}

func (c *Client) GetUsers() ([]user.User, error) {
	glog.Infoln("get mysql users")
	var datas []*User
	//rows, err := c.db.Model(User{}).Select("id, passwd, port, method,enable,transfer_enable,u,d").Rows()
	//查询数据库加上新增的字段
	rows, err := c.db.Model(User{}).Select("id, passwd, port, method,enable,transfer_enable,u,d,end_time,UNIX_TIMESTAMP(LOCALTIME()) as now_time,package_type").Rows()
	if err != nil {
		log.Log.Error(err)
		var users []user.User
		return users, err
	}
	defer rows.Close()
	for rows.Next() {
		var data User
		//err := rows.Scan(&data.id, &data.passwd, &data.port, &data.method, &data.enable, &data.transferEnable, &data.u, &data.d)
		//将读取到的到期时间和账户类型赋值
		err := rows.Scan(&data.id, &data.passwd, &data.port, &data.method, &data.enable, &data.transferEnable, &data.u, &data.d, &data.end_time, &data.now_time, &data.package_type)
		if err != nil {
			log.Log.Error(err)
			continue
		}
		datas = append(datas, &data)
	}
	log.Log.Info(len(datas))
	users := make([]user.User, len(datas))
	for k, v := range datas {
		users[k] = v
	}
	return users, nil
}

func (c *Client) LogNodeOnlineUser(onlineUserCount int) error {
	return nil
}

func (c *Client) UpdateNodeInfo() error {
	return nil
}
