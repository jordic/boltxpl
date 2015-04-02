package boltxpl


var TemplateHTML = `
<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<title>boltXplorer</title>
<meta content='width=device-width,initial-scale=1,maximum-scale=1,user-scalable=no'
	name='viewport'>
<link rel="stylesheet" href="//maxcdn.bootstrapcdn.com/bootstrap/3.3.4/css/bootstrap.min.css" />
<style>
pre {outline: 1px solid #ccc; padding: 5px; margin: 5px; }
.string { color: green; }
.number { color: darkorange; }
.boolean { color: blue; }
.null { color: magenta; }
.key { color: #577492; }

</style>

</head>

<body>

<div class="container">
<div class="row">

	<div class="col-md-12">
		<h3 style="color:#B74934">bolt<span style="color:#577492">xplr</span></h3>
	</div>
</div>

<div class="row">
	<div class="col-md-12" id="table">
		
	</div>
</div>

<div class="row">
	<div class="col-md-12">
		{{ .version }}
	</div>
</div>

</div>



<!-- Modal -->
<div class="modal fade" id="ViewKey" tabindex="-1" role="dialog" aria-labelledby="myModalLabel" aria-hidden="true">
  <div class="modal-dialog">
    <div class="modal-content" id="ViewKeyContent">
    
      
    </div>
  </div>
</div>

<script type="template/javascript" id="tplView">
<div class="modal-header">
        <button type="button" class="close" data-dismiss="modal" aria-label="Close"><span aria-hidden="true">&times;</span></button>
        <h4 class="modal-title" id="myModalLabel"><%= title %></h4>
      </div>
      <div class="modal-body">
        <%= content %>
      </div>
      <div class="modal-footer">
        <button type="button" class="btn btn-default" data-dismiss="modal">Close</button>
      </div>
</script>


<script type="template/javascript" id="tplTable">
<h4><%= p %></h4>
	<table width="100%" class="table">
		<thead>
			<tr>
				<td>Key</td>
				<td>Actions</td>
			</tr>
		</thead>
		<% _.each(c, function(obj, i) { %>
			<tr>
				<td><% if( obj.IsBucket() ) { %>
					<span class="glyphicon glyphicon-folder-open"> </span> &nbsp;<a href="#<%= obj.get("Key") %>" class="bucket"><%= obj.get("Key") %></a>
					<% } else { %>
					<span class="glyphicon glyphicon-file"> </span> &nbsp;<a data-key="<%= obj.get("Key") %>" class="key"><%= obj.get("Key") %></a>
					<% } %>
					</td>
				<td> </td>
			</tr>
			
		<% }) %>
		<tr>
			<td colspan="2" align="right">
			<ul class="pagination">
			<% if (prev != "") { %>
				<li><a href="#<%= p %>?p=<%= prev %>">&laquo; Prev</a></li>
			<% } %>
			<% if (next != "") { %>
				<li><a href="#<%= p %>?p=<%= next %>">Next &raquo;</a></li>
			<% } %>
			</ul>
			</td>		
		</tr>
	</table>
</script>

<script src="//code.jquery.com/jquery-1.11.1.min.js"></script>
<script src="//maxcdn.bootstrapcdn.com/bootstrap/3.3.4/js/bootstrap.min.js"></script>
<script src="//cdnjs.cloudflare.com/ajax/libs/underscore.js/1.8.2/underscore-min.js"></script>
<script src="//cdnjs.cloudflare.com/ajax/libs/backbone.js/1.1.2/backbone-min.js"></script>
<script>

function getQueryVariable(query, variable) {
	if(query.length == 0)
		return ""
    var vars = query.split('&');
    
    for (var i = 0; i < vars.length; i++) {
        var pair = vars[i].split('=');
        if (decodeURIComponent(pair[0]) == variable) {
            return decodeURIComponent(pair[1]);
        }
    }
    return ""
}

function syntaxHighlight(json) {
    if (typeof json != 'string') {
         json = JSON.stringify(json, undefined, 2);
    }
    json = json.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
    return json.replace(/("(\\u[a-zA-Z0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?|\b(true|false|null)\b|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?)/g, function (match) {
        var cls = 'number';
        if (/^"/.test(match)) {
            if (/:$/.test(match)) {
                cls = 'key';
            } else {
                cls = 'string';
            }
        } else if (/true|false/.test(match)) {
            cls = 'boolean';
        } else if (/null/.test(match)) {
            cls = 'null';
        }
        return '<span class="' + cls + '">' + match + '</span>';
    });
}


var app = {}
app.Model = Backbone.Model.extend({
	IsBucket: function() {
		return this.get("IsBucket")
	}
})
app.Keys = Backbone.Collection.extend({
	model: app.Model		
})
app.View = Backbone.View.extend({
	id: 'registros',
	el: '#table',
	
	// Pagination
	perPage : 15,
	currentPage: 0,
	pages : [],
	prev: "",
	next: "",
	Params: "",
	
	Current : "",
	events: {
		"click a.key": "ViewKey" 
	},
	
	ViewKey: function(e) {
		var k = $(e.originalEvent.target).data("key")
		var self = this
		$.get("/-view/", {key: k, bucket: this.Current}, function(data) {
			console.log(data[0])
			if( data[0] == "{" || data[0] == "[" ) {
				var t = JSON.parse(data)
				str = JSON.stringify(t, undefined, 4);
				data = '<pre>' + syntaxHighlight(str) + '</pre>'
				
			} 
		
		
			var html = self.TplView({
				title: k,
				content: data
			})
			$('#ViewKeyContent').html( html )
			$('#ViewKey').modal()
		})
		
		return false;
		
	},
	
	Template: _.template($("#tplTable").html()),
	TplView: _.template($("#tplView").html()),
	
	ResetAndLoadData: function() {
		this.currentPage = 0;
		this.pages = []
		this.next = ""
		this.prev = ""
		this.LoadData()	
	},
	
	LoadData: function() {
		var self = this
		
		if(this.Current == "") {
			url = "/root"
		} else {
			url = "/bucket/" + this.Current
		}
		
		$.get(url, this.GetParams(), function(data) {
			self.collection = new app.Keys(data)
			self.Render()
		})	
	},
	
	GetParams: function() {
		var d = {}
			seek = getQueryVariable(this.Params, "p");
		if(seek != "")
			d["p"] = seek
		return d
	},
	
	Render: function() {
		if(this.collection.length>this.perPage) {
			this.next = this.collection.at(this.perPage).get("Key")
		} else { 
			this.next =""	
		}
		
		var cpos = _.indexOf(this.pages, this.next)
		if( cpos == -1 ) {
			this.pages.push(this.collection.at(0).get("Key"))
		} else {
			this.pages = this.pages.splice(0, cpos)
		}
		
		if( (this.pages.length-2) >= 0)
			this.prev = this.pages[this.pages.length-2]
		else
			this.prev = ""
	
		var html = this.Template({
			p: this.Current,
			c: this.collection.models.slice(0, this.perPage),
			next: this.next,
			prev: this.prev
		})
		$('#table').html( html )
	}
 	
})
app.Router = Backbone.Router.extend({
	routes: {
		"*actions": "default"
	}
});
    
var ins,
	app_router;

$(document).ready(function(){
	ins = new app.View()
	app_router = new app.Router();
	app_router.on('route:default', function(actions) {
		if(!actions) {
			ins.Current	= ""
			ins.Params = ""
			ins.ResetAndLoadData()
		} else {
			var ls = window.location.href; 
			
			if(ls.indexOf("?") != -1) {
				var rest = ls.split("?")
				ins.Params = rest[1]
			} else {
				ins.Params = ""
			}
			
			ins.Current	= actions
			if(ins.Params.length==0) ins.ResetAndLoadData()
			else ins.LoadData()
		}
	})

    // Start Backbone history a necessary step for bookmarkable URL's
    Backbone.history.start();

	
})
</script>
</body>
</html>
`