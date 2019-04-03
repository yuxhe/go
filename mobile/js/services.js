angular.module('chat.services', [])

.factory('Socket', function(socketFactory){
  var opts={};
  opts.transports=['websocket'];
  var myIoSocket = io.connect('192.168.1.106:8088',opts);
  var mySocket = socketFactory({
    ioSocket: myIoSocket
  });
  return mySocket;
})

