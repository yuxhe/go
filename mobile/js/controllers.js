angular.module('chat.controllers', [])

.controller('ChatCtrl', function($scope,$ionicScrollDelegate, $stateParams, $ionicPopup, $timeout, Socket) {
  $scope.username = '';
  $scope.headimg='';
  $scope.accno='';
  $scope.users = {};
  $scope.users.numUsers = 0;
  $scope.isScroll=true;
  $scope.suo=1;	
  $scope.messages = [];
  
  $scope.data = {};
  $scope.data.message = "";

  var typing = false;
  var lastTypingTime;
  var TYPING_MSG = '. . .';

  var Notification = function(room,username,message,img,type,accno){
    var notification          = {};
    notification.room         = room;
    notification.username     = username;
    notification.message      = message;
	notification.headimg      = img;
	notification.type      = type;
    notification.notification = true;
	notification.accno = accno;
    return notification;
  };

  //滚动 
  var scrollBottom = function(){
	if($scope.isScroll){
    	$ionicScrollDelegate.resize();
    	$ionicScrollDelegate.scrollBottom(true);
	}
  };
  
  //加锁
  var noscrollBottom = function(){
	  $scope.isScroll=false;
  };
	//解锁
  var jiesuoscrollBottom = function(){
	  $scope.isScroll=true;
  };
	
  var addMessage = function(msg){
    msg.notification = msg.notification || false;
    $scope.messages.push(msg);
    scrollBottom();
  };

  var setRenshu = function(num){
	  $scope.users.numUsers=num;
  };

  
 var removeTypingMessage = function(usr){
    for (var i = $scope.messages.length - 1; i >= 0; i--) {
      if($scope.messages[i].username === usr && $scope.messages[i].message.indexOf(TYPING_MSG) > -1){
        $scope.messages.splice(i, 1);
        scrollBottom();
        break;
      }
    }
  };

 
  var setAccno =function(acc){
		$scope.accno = acc;
	};
	
	var getRenshu=function(){
		return $scope.users;
	};
	
	var getHeadimg=function(){
      return $scope.headimg;
    };
	
	var setHeadimg=function(himg){
        $scope.headimg = himg;
    };
	
    var  getUsername= function(){
         return $scope.username;
    };
	
    var setUsername=function(usr){
        $scope.username = usr;
    };
    var getMessages=function() {
        return $scope.messages;
    };
	
    var sendMessage=function(msg){
        $scope.messages.push({
        username: $scope.username,
        message:  msg,
		accno :   $scope.accno,
		dateTime: new Date(),
		notification : false,
		type : "sendMessage",
		headimg : $scope.headimg
      });
      scrollBottom();
      Socket.emit('send_message', msg);
    };
	
	var  clearMessages =function(){
	     $scope.messages=[];
		 
 	};

  Socket.on('connect',function(){
    if($stateParams.accno){//!$scope.data.username
	  $scope.data.username=$stateParams.username;
	  $scope.data.room=$stateParams.room;
	  $scope.data.accno=$stateParams.accno;
	  //angular.fromJson(data)   //字符串转对象   对象转字符串 angular.toJson
	  //'{\"Room\":\"room1\",\"Accno\":\"accno2\"}'
	  Socket.emit('join_room',angular.toJson({Room:$scope.data.room,Accno:$scope.data.accno}));
      //屏蔽掉 通过服务端转发
	  //Socket.emit('joined_message',$scope.data.username);
      setUsername($scope.data.username);
	  setAccno($scope.data.accno);
	  $scope.messages=[];
	
	  /*
	  for  (var i=0;i<tmp.length;i++) {
		    $scope.messages.push(tmp[i]);
	  }
	  */
	
	  /* //2016-1-12 yuxh 屏蔽掉 下面代码，采用上面动态机制处理代码
      var nicknamePopup = $ionicPopup.show({
      template: '<input id="usr-input" type="text" ng-model="data.username" autofocus>',
      title: 'What\'s your nickname?',
      scope: $scope,
      buttons: [{
          text: '<b>Save</b>',
          type: 'button-positive',
          onTap: function(e) {
            if (!$scope.data.username) {
              e.preventDefault();
            } else {
              return $scope.data.username;
            }
          }
        }]
      });
	  
      nicknamePopup.then(function(username) {
		Socket.emit('join_room','{\"Room\":\"room1\",\"Accno\":\"accno2\"}')
        Socket.emit('joined_message',username);
        Chat.setUsername(username);
      });
	  */
	
    }

  });


  //----------------------------------------
  Socket.on('message', function(data){
	  var msg = angular.fromJson(data)
      addMessage(msg);
  });

  Socket.on('typing', function (data) {
    var typingMsg = {
      username: data.username,
      message: TYPING_MSG
    };
     addMessage(typingMsg);
  });

  Socket.on('stop typing', function (data) {
    removeTypingMessage(data.username);
  });

  Socket.on('room joined', function (data) {
	data=angular.fromJson(data);
    var msg = data.room + ' joined';
	setRenshu(data.numUsers);
	$scope.username = data.username;
	$scope.headimg = data.headimg;
	$scope.accno = data.accno;
	//console.log($scope.headimg +"  " +$scope.accno)
    addMessage(msg);
  });

  Socket.on('user joined', function (data) {
	var msg = angular.fromJson(data)
    var notification = new Notification('',msg.username,msg.username + " 加入【" + msg.gpname + "】聊天室",msg.headimg,msg.type,msg.accno);
    addMessage(notification);
	setRenshu(msg.numUsers);
  });

  Socket.on('user left', function (data) {
	data=angular.fromJson(data);
    var msg = data.username + ' 离开聊天室('+data.dateTime+')';
    var notification = new Notification('',data.username,msg,'');
    addMessage(notification);
	setRenshu(data.numUsers);
  });
  //----------------------------------------

  scrollBottom();

  $scope.mobileType = function(){
	var mtype=$stateParams.mobile;
	if(mtype !=null && mtype != ""){
		if(mtype == "android"){
			//console.log("androidContent");
			return "androidContent";
		}else{
			return "iphoneContent";
		}
	}else{
		return "iphoneContent";
	}
   }
 $scope.mobileTypeTOP = function(){
	var mtype=$stateParams.mobile;
	if(mtype !=null && mtype != ""){
		if(mtype == "android"){
			//console.log("androidContentTop");
			return "androidContentTop";
		}else{
			return "iphoneContentTop";
		}
	}else{
		return "iphoneContentTop";
	}
   }


  if($stateParams.username){
    $scope.data.message ="" //"@" + $stateParams.username;
    document.getElementById("msg-input").focus();
  } 

  var sendUpdateTyping = function(){
    if (!typing) {
      typing = true;
      Socket.emit('typing');
    }
	var TYPING_TIMER_LENGTH=500;
    lastTypingTime = (new Date()).getTime();
    $timeout(function () {
      var typingTimer = (new Date()).getTime();
      var timeDiff = typingTimer - lastTypingTime;
      if (timeDiff >= TYPING_TIMER_LENGTH && typing) {
        Socket.emit('stop typing');
        typing = false;
      }
    }, TYPING_TIMER_LENGTH);
  };

  $scope.updateTyping = function(){
      sendUpdateTyping();
  };

  $scope.messageIsMine = function(accno){
    return $scope.data.accno === accno;
  };


  $scope.getBubbleClass = function(accno){
    var classname = 'from-them';
    if($scope.messageIsMine(accno)){
      classname = 'from-me';
    }
    return classname;
  };

  $scope.getUnClass = function(accno){
    var classname = 'username';
    if($scope.messageIsMine(accno)){
      classname = 'username from-unme';
    }
    return classname;
  };

  $scope.sendMessage = function(msg){
	if(msg != null && msg != ""){
		sendMessage(msg);
	    $scope.data.message = "";
		//发送后自动解锁
		$scope.suo=1;
		jiesuoscrollBottom();
		scrollBottom();
	}
   
	
  };

  $scope.clearMessage = function(){
	clearMessages();
	$scope.suo=1;
    jiesuoscrollBottom();
    scrollBottom();
  };

 $scope.isJoinMessage = function(type,headimg){
	//headimg="http://192.168.1.14/mediafiles/media?uuid=";
	var c;
	
	if(headimg !=null && headimg !=""){
		var ch=headimg.substring(headimg.length-1,headimg.length);
		//console.log(ch);
			if(ch != "="){
				if(type !=null && type != ""){
				if(type == "joined_message" || type == "joined_room"){
					c=false;
				}else{
					c=true;
				}		
			}else{
				c=false;
			}
		}else{
			c=false;
		}	
	}else{
		c=false;
	}
	return c;
	};
	
	$scope.isJoin2Message = function(type,headimg){
		var c;
		if(headimg !=null && headimg !=""){
			var ch=headimg.substring(headimg.length-1,headimg.length);
			if(type !=null && type != ""){
				if(type == "joined_message" || type == "joined_room"){
						c=false;
				}else{
					if(ch != "="){
						c=false;
					}else{
						c=true;
					}
				}
			}
		}else{
			c=false;
		}
		return c;
	}
	
//人数
  $scope.renshu = getRenshu();

  //锁
  $scope.lockMessage=function(){
  $scope.suo=2;
  noscrollBottom();
  };

  $scope.solutionLockMessage=function(){
  $scope.suo=1;
  jiesuoscrollBottom();
  scrollBottom();
  };
   
 $scope.closeWindow = function(){
	window.opener = null;
	window.open("","_self");
	window.close();
 }	
	
})


/*
.controller('PeopleCtrl', function($scope, Users) {
    $scope.data = Users.getUsers();
	$scope.userList = Users.getUsersList();
})
*/
