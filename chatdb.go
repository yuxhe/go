package main

import (
	MongoUtil "mychatapp/mongoutil"
	"time"

	"gopkg.in/mgo.v2/bson"
)

/*
**   getSystemParam 获取系统参数值
 */
func getSystemParam(db MongoUtil.Db, syscode string) string {
	base := MongoUtil.BaseRepository{db, "sys_parameter"}
	var p bson.M
	p = base.FindOne(bson.M{"status": 0, "sysparamcode": syscode},
		[]string{}, bson.M{})
	if p["sysparamvalue"] == nil {
		return ""
	} else {
		return p["sysparamvalue"].(string)
	}
}

/*
**   getRoomAccno 核查房间号(圈子号)同 accno的对应
 */
func getRoomAccno(db MongoUtil.Db, room string, accno string, file_url string) bson.M {
	//1、依据 groupno 查找 gpid
	base := MongoUtil.BaseRepository{db, "gup_groupinfo"}
	var p bson.M
	var p_member bson.M
	var gpname string
	var headimg string
	var memname string
	p = base.FindOne(bson.M{"isdelete": 0, "groupno": room},
		[]string{}, bson.M{})
	if p["_id"] == nil {
		return bson.M{}
	} else {
		//2、查找头像
		base_member := MongoUtil.BaseRepository{db, "mem_memberinfo"}
		p_member = base_member.FindOne(bson.M{"accno": accno},
			[]string{}, bson.M{})
		headimg = file_url + p_member["headimg"].(string) //头像
		memname = p_member["memname"].(string)            //用户名称

		base := MongoUtil.BaseRepository{db, "gup_groupmember"}
		//2、依据  gpid 查找对应的用户
		//3、加入用户名 及用户的头像

		gpname = p["gpname"].(string)
		p = base.FindOne(bson.M{"gpmtype": 9, "gpid": p["_id"].(bson.ObjectId).Hex(), "gpmno": accno},
			[]string{}, bson.M{})
		p["gpname"] = gpname
		p["username"] = memname
		p["headimg"] = headimg
		return p
	}
}

/*
**   getRoomAccno 核查房间是否存在，若不存在则创建房间
 */
func getChatRoom(db MongoUtil.Db, groupno string, accno string) string {
	base := MongoUtil.BaseRepository{db, "cht_chatroom"}
	var p bson.M
	p = base.FindOne(bson.M{"croomtype": 1, "isdelete": 0, "groupno": groupno},
		[]string{}, bson.M{})

	if p["groupno"] == nil {
		//插入房间,返回房间的_id
		id := base.Insert(bson.M{"croomtype": 1, "croomno": groupno, "groupno": groupno,
			"isdelete": 0, "createuser": accno, "createtime": time.Now().Format("2006-01-02 15:04:05"),
			"lupdateuser": accno, "lupdatetime": time.Now().Format("2006-01-02 15:04:05")})
		return id
	} else {
		return p["_id"].(bson.ObjectId).Hex()
	}
}

/*
**   insertChatMsg  插入聊天信息
 */
func insertChatMsg(db MongoUtil.Db, croomid string, accno string, msgtxt string, msgtype int) string {
	var lastid = Substr(croomid, len(croomid)-1, 1)
	base := MongoUtil.BaseRepository{db, "cht_chatmsg_" + lastid}
	var id = base.Insert(bson.M{"croomid": croomid, "accno": accno, "msgtxt": msgtxt, "msgtype": 1, "msgdate": time.Now().Format("2006-01-02 15:04:05")})
	return id
}

/*
**  Substr 截取字符串  start 开始 length 多长
 */
func Substr(str string, start, length int) string {
	rs := []rune(str)
	rl := len(rs)
	end := 0

	if start < 0 {
		start = rl - 1 + start
	}
	end = start + length

	if start > end {
		start, end = end, start
	}

	if start < 0 {
		start = 0
	}
	if start > rl {
		start = rl
	}
	if end < 0 {
		end = 0
	}
	if end > rl {
		end = rl
	}
	return string(rs[start:end])
}
