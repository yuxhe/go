package main

import (
	"encoding/json"
	"log"
	MongoUtil "mychatapp/mongoutil"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/googollee/go-socket.io"
	_ "gopkg.in/mgo.v2/bson"
)

// Should look like path
const websocketRoom = "/chat"

type Roomuser struct {
	Room  string
	Accno string
}

func main() {
	//-----------------------------------------
	var db = MongoUtil.Db{"datebase_luyoutcore", "luyoutcore"} //数据库连接的配置 及选择数据库
	/*
		base := MongoUtil.BaseRepository{db, "person"}

		var p bson.M
		p = base.FindOne(bson.M{"Name": "zhangsan2", "Phone": "028-3338"},
			[]string{}, bson.M{"Name": 1})
		if p == nil {
			log.Println("person", 123213)
		}
		log.Println("person", p["_id"].(bson.ObjectId).Hex()) //reflect.TypeOf(p["_id"])

		return
	*/

	/*
		ttt := base.Update(bson.M{"Name": "zhangsan2", "Phone": "028-3338"},
				bson.M{"Name": "zhangsan21", "Phone": "028-3339"})
		log.Println("person", ttt)
		return
	*/

	//
	//-----------------------------------------
	//------------------------获取系统参数 file_url路径
	var file_url = getSystemParam(db, "file_url")
	//-----------------------------------------
	lastMessages := make(map[string][]string)
	numUsers := make(map[string]int)                  //每个房间的在线人数
	numUsersOnline := make(map[string]map[string]int) //在线的注册用户
	var lmMutex sync.Mutex
	// Sets the number of maxium goroutines to the 2*numberCPU + 1
	runtime.GOMAXPROCS((runtime.NumCPU() * 2) + 1)

	// Configuring socket.io Server
	sio, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
	}
	sio.SetMaxConnection(20000) // add yuxh on 2016-3-2 默认为1000连接 暂时设置最大20000个连接
	sio.On("connection", func(so socketio.Socket) {
		var username string //用户名称
		var headimg string  //头像
		var room string     //对应groupid
		var accno string    //对应accno
		var chtmsgid string //房间的chtmsgid
		var gpname string   //房间的中文名称

		var valid bool = false //不合法用户
		username = "User-" + so.Id()
		log.Println("on connection", username)

		so.On("join_room", func(message string) {
			log.Println("开始房间", message)
			//-----------------接收圈子号  room 及 accno
			roomuser := &Roomuser{}
			err = json.Unmarshal([]byte(message), &roomuser)
			log.Println("error:", err)
			room = roomuser.Room
			accno = roomuser.Accno
			//-----------------1、判断房间room  及 accno的对应关系
			//核查accno是否属于XXXX房间
			//-------------------------------------------------------
			/*
				base := MongoUtil.BaseRepository{db, "gup_groupmember"}
				var p bson.M
				p = base.FindOne(bson.M{"gpmtype": 9, "gpid": room, "gpmno": accno},
					[]string{}, bson.M{})
			*/
			p := getRoomAccno(db, room, accno, file_url) //room 为圈子标识号
			if p["_id"] == nil {
				valid = false //非正常用户
				log.Println("用户状态", "无效用户:"+accno)
				return
			} else {
				valid = true //标记为合法用户
			}
			gpname = p["gpname"].(string)     //加入的房间号名称
			username = p["username"].(string) //用户名称
			headimg = p["headimg"].(string)   //用户头像
			log.Println("合法用户", p["gpmno"])
			//-------------------------------------------------------
			//核查是否有房间，无则房间创建
			chtmsgid = getChatRoom(db, room, accno)
			//-------------------------------------------------------
			//room = message  暂时屏蔽掉
			log.Println("join_room", room)
			log.Println("join_Accno", accno)
			so.Join(room)
			//---------------------------------------处理在线人数
			lmMutex.Lock()
			onlineroom, ok := numUsersOnline[room]
			if ok {
				_, ok := onlineroom[accno]
				if !ok {
					onlineroom[accno] = 1
					numUsers[room]++
				} else {
					onlineroom[accno] = onlineroom[accno] + 1
				}
			} else {
				//该房间一个人都没有的情况
				usrtmp := make(map[string]int)
				usrtmp[accno] = 1
				numUsersOnline[room] = usrtmp
				numUsers[room]++
			}
			//------------------------------------

			for i, _ := range lastMessages[room] {
				so.Emit("message", lastMessages[room][i])
			}
			lmMutex.Unlock()

			res := map[string]interface{}{
				"room":     "join " + room,
				"username": username,
				"headimg":  headimg,
				"accno":    accno,
				"numUsers": numUsers[room],
				"dateTime": time.Now().Format("2006-01-02 15:04:05"),
				"type":     "joined_room",
			}
			jsonRes, _ := json.Marshal(res)
			so.Emit("room joined", string(jsonRes))
			so.BroadcastTo(room, "message", string(jsonRes))
			//---------------------------------------试试 优化直接服务端 转化圈子的名称
			//???
			var accnonum int = 0
			var v_jsonRes []byte
			joined_message := func() {
				res := map[string]interface{}{
					"username": username,
					"headimg":  headimg,
					"room":     room,
					"accno":    accno,
					"numUsers": numUsers[room],
					"gpname":   gpname,
					"dateTime": time.Now().Format("2006-01-02 15:04:05"),
					"type":     "joined_message",
				}
				v_jsonRes, _ = json.Marshal(res)
				so.Emit("user joined", string(v_jsonRes))
				lmMutex.Lock()
				accnonum, _ = numUsersOnline[room][accno]
				lmMutex.Unlock()
			}
			//log.Println("验证转发开始时间")
			joined_message() //直接转发信息了
			//???
			//log.Println("验证广播开始时间")
			if valid && accnonum == 1 { //合法用户才发起广播到其他的用户
				so.BroadcastTo(room, "user joined", string(v_jsonRes))
			}
			log.Println("直接转发完成")
			//---------------------------------------
		})

		/*
			so.On("joined_message", func(message string) {
				//username = message   使用系统取出的用户名
				log.Println("joined_message", username)
				log.Println("joined_message", headimg)

				if !valid {
					username = "非法账户:" + username + " 不能参与正常聊天"
				}
				res := map[string]interface{}{
					"username": username,
					"headimg":  headimg,
					"room":     room,
					"accno":    accno,
					"numUsers": numUsers[room],
					"gpname":   gpname,
					"dateTime": time.Now().Format("2006-01-02 15:04:05"),
					"type":     "joined_message",
				}
				jsonRes, _ := json.Marshal(res)
				so.Emit("user joined", string(jsonRes))
				lmMutex.Lock()
				accnonum, _ := numUsersOnline[room][accno]
				if valid && accnonum == 1 { //合法用户才发起广播到其他的用户
					so.BroadcastTo(room, "user joined", string(jsonRes))
				}
				lmMutex.Unlock()
			})
		*/

		so.On("left_message", func(message string) {
			//username = message
			log.Println("left_message", message)
			lmMutex.Lock()
			accnonum, ok := numUsersOnline[room][accno]
			if ok {
				if accnonum == 1 {
					delete(numUsersOnline[room], accno)
					numUsers[room]--
				} else {
					numUsersOnline[room][accno]--
				}
			}
			lmMutex.Unlock()
			res := map[string]interface{}{
				"username": username,
				"headimg":  headimg,
				"room":     room,
				"accno":    accno,
				"numUsers": numUsers[room],
				"dateTime": time.Now().Format("2006-01-02 15:04:05"),
				"type":     "left_message",
			}
			jsonRes, _ := json.Marshal(res)
			so.Emit("user left", string(jsonRes))
			so.BroadcastTo(room, "message", string(jsonRes))
		})
		so.On("send_message", func(message string) {
			log.Println("send_message from", username)
			res := map[string]interface{}{
				"username": username,
				"headimg":  headimg,
				"room":     room,
				"accno":    accno,
				"message":  message,
				"dateTime": time.Now().Format("2006-01-02 15:04:05"),
				"type":     "message",
			}
			jsonRes, _ := json.Marshal(res)
			lmMutex.Lock()
			//调整代码 从原来的100 条改为最近的10条
			if len(lastMessages[room]) == 10 {
				lastMessages[room] = lastMessages[room][1:10]
			}
			lastMessages[room] = append(lastMessages[room], string(jsonRes))
			lmMutex.Unlock()
			//so.Emit("message", string(jsonRes))
			if valid { //合法用户正常转发信息
				so.BroadcastTo(room, "message", string(jsonRes))
				//插入聊天记录
				insertChatMsg(db, chtmsgid, accno, message, 1)
			}
		})
		so.On("disconnection", func() {
			log.Println("on disconnect", username)
			lmMutex.Lock()
			accnonum, ok := numUsersOnline[room][accno]
			if ok {
				if accnonum == 1 {
					delete(numUsersOnline[room], accno)
					numUsers[room]--
				} else {
					numUsersOnline[room][accno]--
				}
			}
			lmMutex.Unlock()
			res := map[string]interface{}{
				"username": username,
				"headimg":  headimg,
				"room":     room,
				"accno":    accno,
				"numUsers": numUsers[room],
				"dateTime": time.Now().Format("2006-01-02 15:04:05"),
				"type":     "left_message",
			}
			jsonRes, _ := json.Marshal(res)
			if accnonum == 1 {
				so.BroadcastTo(room, "user left", string(jsonRes))
			}
		})
	})
	sio.On("error", func(so socketio.Socket, err error) {
		log.Println("error:", err)
	})

	// Sets up the handlers and listen on port 8080
	http.Handle("/socket.io/", sio)

	http.Handle("/lib/", http.StripPrefix("/lib/", http.FileServer(http.Dir("./mobile/lib/"))))
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("./mobile/css/"))))
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("./mobile/js/"))))

	//http.Handle("/", http.FileServer(http.Dir("./mobile/")))
	http.Handle("/mobile/", http.StripPrefix("/mobile", http.FileServer(http.Dir("./mobile/"))))

	http.Handle("/impc/", http.StripPrefix("/impc", http.FileServer(http.Dir("./pc/"))))

	// Default to :8080 if not defined via environmental variable.
	var listen string = os.Getenv("LISTEN")

	if listen == "" {
		listen = ":8088"
	}

	log.Println("listening on", listen)
	http.ListenAndServe(listen, nil)
}
