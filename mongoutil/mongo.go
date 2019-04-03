package mongoutil

import (
	_ "encoding/json"
	_ "fmt"
	"reflect"
	"strings"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

/*
type Person struct {
	Id    bson.ObjectId `bson:"_id"`
	Name  string        `bson:"Name"` //bson:"name" 表示mongodb数据库中对应的字段名称
	Phone string        `bson:"Phone"`
}
*/

/*
 * BaseRepository
 */

type BaseRepository struct {
	//URL        string `mongodb://luyoutcore:luyoutcore123@192.168.1.13:21001,192.168.1.14:21001/luyoutcore`
	Dbln       Db
	Collection string `Base` //操作的集合
}

var (
	cf         Config
	mgoSession *mgo.Session
)

/**
*  公共方法，0、初始化配置文件读取
 */
func init() {
	cf = Config{}                      //new(Config)
	cf.InitConfig("config.properties") //配置文件初始化读取
}

/**
 * 公共方法，1、获取连接session
 */
func (col BaseRepository) getSession() *mgo.Session {
	if mgoSession == nil {
		var err error
		mgoSession, err = mgo.Dial(cf.Get(col.Dbln.URL)) //URL 配置
		if err != nil {
			panic(err) //直接终止程序运行
		}
	}
	//最大连接池默认为4096
	return mgoSession.Clone()
}

/**
 * 公共方法，2、获取待操作的collection
 */
func (col BaseRepository) witchCollection(s func(*mgo.Collection) error) error {
	session := col.getSession()
	defer session.Close()
	c := session.DB(col.Dbln.DataBase).C(col.Collection)
	return s(c)
}

/**
 * 公共方法，3、Insert 插入记录
 */
func (col BaseRepository) Insert(bsonM bson.M) string {
	objid := bson.NewObjectId()
	bsonM["_id"] = objid
	query := func(c *mgo.Collection) error {
		return c.Insert(bsonM)
	}
	err := col.witchCollection(query)
	if err != nil {
		return "0"
	}
	return objid.Hex()
}

/**
 * 公共方法，4、GetDataById 获取单条数据按id查询
 */
func (col BaseRepository) GetDataById(id string, fields bson.M) bson.M {
	objid := bson.ObjectIdHex(id)
	bsonM := bson.M{}
	query := func(c *mgo.Collection) error {
		return c.FindId(objid).Select(fields).One(bsonM)
	}
	col.witchCollection(query)
	return bsonM
}

/**
 * 公共方法，5、FindOne 按条件查询单条数据 ,查询条件、排序条件、字段过滤
 */
func (col BaseRepository) FindOne(query bson.M, sorts []string, fields bson.M) bson.M {
	bsonM := bson.M{}
	exop := func(c *mgo.Collection) error {
		return c.Find(query).Sort(sorts...).Select(fields).One(bsonM)
	}
	col.witchCollection(exop)
	return bsonM
}

/**
 * 公共方法，6、FindList 不分页的查询 按条件查询数据 ,查询条件、排序条件、字段过滤
 */
func (col BaseRepository) FindList(query bson.M, sorts []string, fields bson.M) []bson.M {
	bsonML := []bson.M{}
	exop := func(c *mgo.Collection) error {
		return c.Find(query).Sort(sorts...).Select(fields).All(&bsonML)
	}
	col.witchCollection(exop)
	return bsonML
}

/**
 * 公共方法，7、FindPageList 分页查询按条件查询数据 ,查询条件、排序条件、字段过滤
 */
func (col BaseRepository) FindPageList(query bson.M, sorts []string, fields bson.M, lastitemid string, page int, rows int) []bson.M {
	if page > 0 {
		page = page - 1
	} else {
		page = 0
	}
	bsonML := []bson.M{}
	if lastitemid != "" && (!("" == strings.TrimSpace(lastitemid))) {
		exop := func(c *mgo.Collection) error {
			return c.Find(query).Sort(sorts...).Select(fields).Limit(rows).All(&bsonML)
		}
		col.witchCollection(exop)
	} else {
		exop := func(c *mgo.Collection) error {
			return c.Find(query).Sort(sorts...).Select(fields).Limit(rows).Skip(page * rows).All(&bsonML)
		}
		col.witchCollection(exop)
	}

	return bsonML
}

/**
 * 公共方法，8、update 修改数据
 */
func (col BaseRepository) Update(query bson.M, change bson.M) string {
	AddBool := true
	for key, _ := range change {
		//fmt.Println(reflect.TypeOf(key))  反射机制的运用 能递归 遍历
		if reflect.TypeOf(key).String() == "string" {
			if strings.HasPrefix(reflect.ValueOf(key).String(), "$") {
				AddBool = false
				break
			}
		}
	}

	exop := func(c *mgo.Collection) error {
		if AddBool {
			return c.Update(query, bson.M{"set": change})
		} else {
			return c.Update(query, change)
		}
	}
	err := col.witchCollection(exop)
	if err != nil {
		return "false"
	}
	return "true"
}

/**
 * 公共方法，9、deleteById 按id删除数据
 */
func (col BaseRepository) deleteById(id string) string {
	objid := bson.ObjectIdHex(id)
	query := func(c *mgo.Collection) error {
		return c.RemoveId(objid)
	}
	err := col.witchCollection(query)
	if err != nil {
		return "false"
	}
	return "true"
}

/**
 * 执行查询，此方法可拆分做为公共方法
 * [SearchPerson description]
 * @param {[type]} collectionName string [description]
 * @param {[type]} query          bson.M [description]
 * @param {[type]} sort           bson.M [description]
 * @param {[type]} fields         bson.M [description]
 * @param {[type]} skip           int    [description]
 * @param {[type]} limit          int)   (results      []interface{}, err error [description]
 */
/*
func SearchPerson(collectionName string, query bson.M, sort string, fields bson.M, skip int, limit int) (results []interface{}, err error) {
	exop := func(c *mgo.Collection) error {
		c.Find(query).se
		return c.Find(query).Sort(sort).Select(fields).Skip(skip).Limit(limit).All(&results)
	}
	err = witchCollection(collectionName, exop)
	return
}
*/

func main() {
	/*
		p := Person{}
		p.Name = "zhangsan1"
		p.Phone = "028-3334"
		AddPerson(p)
	*/

	/*
		var p *Person
		p = GetPersonById("568dd0172f50c42294333d61")
	*/

	/*
		var p []Person
		p = PagePerson()
	*/

	/*
		var p bson.M
			p = FindOne(bson.M{"Name": "zhangsan1", "Phone": "028-3334"})
			fmt.Println(reflect.TypeOf(p))
			//fmt.Println(p)
			b, err := json.Marshal(p)
			if err != nil {
				fmt.Println("error:", err)
			}
			fmt.Println(string(b))
	*/

	/*
		var p bson.M
		baseperson := &BaseRepository{"person"}
		p = baseperson.FindOne(bson.M{"Name": "zhangsan2", "Phone": "028-3338"})
		fmt.Println(p)
	*/

	/*
		ttt := Update(bson.M{"Name": "zhangsan1", "Phone": "028-3334"}, bson.M{"$set": bson.M{"Name": "zhangsan2", "Phone": "028-3338"}})
		fmt.Println(ttt)
	*/

	/*
		var p bson.M
		p = FindOne(bson.M{"Name": "zhangsan1", "Phone": "028-3334"})
		fmt.Println(reflect.TypeOf(p))
		//fmt.Println(p)
		b, err := json.Marshal(p)
		if err != nil {
			fmt.Println("error:", err)
		}
		fmt.Println(string(b))
	*/

	/*  处理返回多条的情况
	var p []bson.M
	p = FindList(bson.M{"Name": "zhangsan1", "Phone": "028-3334"})
	fmt.Println(reflect.TypeOf(p))
	//fmt.Println(p)
	b, err := json.Marshal(p)
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Println(string(b))
	*/

	//p := bson.M{"tname": "zhangsan", "tphone": "028-3333"}

	/*
		for key, value := range p {
			//fmt.Println(reflect.TypeOf(key))  反射机制的运用 能递归 遍历
			fmt.Println(reflect.TypeOf(p))
			fmt.Println(key)
			fmt.Println(value)
		}
	*/

	//fmt.Print(p["tname"])

	//tt := AddPerson(p)
	//fmt.Print(tt)
}
