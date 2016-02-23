var channel = "#default";
var url = window.location.protocol + "//" + window.location.host + window.location.pathname;
var wsurl = wsURL();
var ws = new WebSocket(wsurl);

function wsURL() {
	var u = "ws://" + window.location.host + window.location.pathname

	if (window.location.pathname.slice(-1) != "/") {
		u += "/";
	}

	return u + "stream";
}

function changed() {
    if (window.location.hash.length > 0) {
        channel = window.location.hash;
    } else {
        channel = "#default";
    }
    read();
    stream();
};

function recv(ev) {
    var data = JSON.parse(ev.data)
    var date = new Date(data.created/1e6);

    item = "<div id='" + data.id + "'>"
	+ "<div>"
	+ "<b>" + data.from + "</b>" + " " + date.toLocaleTimeString()
	+ "</div>"
	+ data.text
	+ "</div>";

    $('.screen').append(item);
        var d = $('.screen');
        d.scrollTop(d.prop("scrollHeight"));
}

function read() {
    $.ajax({
      dataType: "json",
      url: url + '/read',
      data: {
        channel: channel,
      },
      success: function(data) {
        $('.channel').text(channel);

        var items = [];
        $.each(data["events"], function(key, val) {
            var date = new Date(val.created/1e6);
            items.push( "<div id='" + val.id + "'>"
                + "<div>"
                + "<b>" + val.from + "</b>" + " " + date.toLocaleTimeString()
                + "</div>"
		+ val.text
		+ "</div>"
	    );
        })
        $('.screen').empty();
        $('.screen').append(items.join(""));
        var d = $('.screen');
        d.scrollTop(d.prop("scrollHeight"));
      }
    });
};

function write(text, fn) {
    $.ajax({
      dataType: "json",
      url: url + '/write',
      data: {
        channel: channel,
        text: text,
        from: $('#username').val(),
      },
      success: fn,
    });
}; 

function stream() {
    ws.onclose = function () {};
    ws.close();
    ws = new WebSocket(wsurl);
    ws.onmessage = recv;
    ws.onopen = function (event) {
      ws.send(JSON.stringify({
          channel: channel,
      }));
    }
};

function username() {
    $('#username').val(localStorage.username);

    var val = $('#username').val();

    if (val.length == 0) {
	$.ajax({
	  url: 'https://randomuser.me/api/',
	  dataType: 'json',
	  success: function(data){
            $('#username').val(data.results[0].user.name.first);
            localStorage.username = data.results[0].user.name.first;
	  }
	});
    };
};

function load() {
    read();
    stream();
    username();

    $('#text-form').submit(function(e) {
        e.preventDefault();
        text = $(this).serializeArray()[0].value;
        write(text, function() {
	    $('#text').val('');
	});
    });

    $('#username-form').submit(function(e) {
        e.preventDefault();
        localStorage.username = $('#username').val();
	return;
    });
};

$(document).ready(function() {
    load();
});
