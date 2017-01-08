/* This file was automatically generated by the software. Please don't edit it as your changes may be overwritten at any moment. */
package main
import "strconv"
import "html/template"
import "io"

func init() {
	template_topic_alt_handle = template_topic_alt
	//o_template_topic_alt_handle = template_topic_alt
	ctemplates = append(ctemplates,"topic_alt")
	tmpl_ptr_map["topic_alt"] = &template_topic_alt_handle
	tmpl_ptr_map["o_topic_alt"] = template_topic_alt
}

func template_topic_alt(tmpl_topic_alt_vars TopicPage, w io.Writer) {
w.Write([]byte(`<!doctype html>
<html lang="en">
	<head>
		<title>` + tmpl_topic_alt_vars.Title + `</title>
		<link href="/static/main.css" rel="stylesheet" type="text/css">
		<script type="text/javascript" src="/static/jquery-1.12.3.min.js"></script>
		<script type="text/javascript">
		var session = "` + tmpl_topic_alt_vars.CurrentUser.Session + `";
		</script>
		<script type="text/javascript" src="/static/global.js"></script>
		<meta name="viewport" content="width=device-width,initial-scale = 1.0, maximum-scale=1.0,user-scalable=no" />
	</head>
	<body>
		<div class="container">
<div class="nav">
	<div class="move_left">
	<div class="move_right">
	<ul>
		<li class="menu_left menu_overview"><a href="/">Overview</a></li>
		<li class="menu_left menu_forums"><a href="/forums/">Forums</a></li>
		<li class="menu_left menu_topics"><a href="/">Topics</a></li>
		<li class="menu_left menu_create_topic"><a href="/topics/create/">Create Topic</a></li>
		`))
if tmpl_topic_alt_vars.CurrentUser.Loggedin {
w.Write([]byte(`
		<li class="menu_left menu_account"><a href="/user/edit/critical/">Account</a></li>
		<li class="menu_left menu_profile"><a href="/user/` + strconv.Itoa(tmpl_topic_alt_vars.CurrentUser.ID) + `">Profile</a></li>
		`))
if tmpl_topic_alt_vars.CurrentUser.Is_Super_Mod {
w.Write([]byte(`<li class="menu_left menu_account"><a href="/panel/">Panel</a></li>`))
}
w.Write([]byte(`
		<li class="menu_left menu_logout"><a href="/accounts/logout?session=` + tmpl_topic_alt_vars.CurrentUser.Session + `">Logout</a></li>
		`))
} else {
w.Write([]byte(`
		<li class="menu_left menu_register"><a href="/accounts/create/">Register</a></li>
		<li class="menu_left menu_login"><a href="/accounts/login/">Login</a></li>
		`))
}
w.Write([]byte(`
	</ul>
	</div>
	</div>
	<div style="clear: both;"></div>
</div>
<div id="back"><div id="main">`))
if len(tmpl_topic_alt_vars.NoticeList) != 0 {
for _, item := range tmpl_topic_alt_vars.NoticeList {
w.Write([]byte(`<div class="alert">` + item + `</div>`))
}
}
w.Write([]byte(`
<div class="rowblock">
	<form action='/topic/edit/submit/` + strconv.Itoa(tmpl_topic_alt_vars.Topic.ID) + `' method="post">
		<div class="rowitem rowhead`))
if tmpl_topic_alt_vars.Topic.Sticky {
w.Write([]byte(` topic_sticky_head`))
} else {
if tmpl_topic_alt_vars.Topic.Is_Closed {
w.Write([]byte(` topic_closed_head`))
}
}
w.Write([]byte(`">
			<a class='topic_name hide_on_edit'>` + tmpl_topic_alt_vars.Topic.Title + `</a> 
			`))
if tmpl_topic_alt_vars.Topic.Is_Closed {
w.Write([]byte(`<span class='username hide_on_micro topic_status_e topic_status_closed hide_on_edit' title='Status: Closed' style="font-weight:normal;float: right;">&#x1F512;&#xFE0E</span>`))
}
w.Write([]byte(`
			`))
if tmpl_topic_alt_vars.CurrentUser.Is_Mod {
w.Write([]byte(`
			<a href='/topic/edit/` + strconv.Itoa(tmpl_topic_alt_vars.Topic.ID) + `' class="username hide_on_edit open_edit topic_button" style="font-weight: normal;margin-left: 6px;">Edit</a>
			<a href='/topic/delete/submit/` + strconv.Itoa(tmpl_topic_alt_vars.Topic.ID) + `' class="username topic_button" style="font-weight: normal;">Delete</a>
			`))
if tmpl_topic_alt_vars.Topic.Sticky {
w.Write([]byte(`<a href='/topic/unstick/submit/` + strconv.Itoa(tmpl_topic_alt_vars.Topic.ID) + `' class="username topic_button" style="font-weight: normal;">Unpin</a>`))
} else {
w.Write([]byte(`<a href='/topic/stick/submit/` + strconv.Itoa(tmpl_topic_alt_vars.Topic.ID) + `' class="username topic_button" style="font-weight: normal;">Pin</a>`))
}
w.Write([]byte(`
			
			<input class='show_on_edit topic_name_input' name="topic_name" value='` + tmpl_topic_alt_vars.Topic.Title + `' type="text" />
			<select name="topic_status" class='show_on_edit topic_status_input' style='float: right;'>
				<option>open</option>
				<option>closed</option>
			</select>
			<button name="topic-button" class="formbutton show_on_edit submit_edit">Update</button>
			`))
}
w.Write([]byte(`
			<a href="/report/submit/` + strconv.Itoa(tmpl_topic_alt_vars.Topic.ID) + `?session=` + tmpl_topic_alt_vars.CurrentUser.Session + `&type=topic" class="username report_item topic_button" style="font-weight: normal;">Report</a>
		</div>
	</form>
</div>
<style type="text/css">.rowitem:last-child .content_container { margin-bottom: 5px !important; }</style>
<div class="rowblock post_container" style="border-top: none;">
	<div class="rowitem passive deletable_block editable_parent post_item" style="background-color: #eaeaea;padding-top: 4px;padding-left: 5px;clear: both;border-bottom: none;padding-right: 4px;padding-bottom: 2px;">
		<div class="userinfo">
			<div class="avatar_item" style="background-image: url(` + tmpl_topic_alt_vars.Topic.Avatar + `), url(/static/white-dot.jpg);background-position: 0px -10px;">&nbsp;</div>
			<a href="/user/` + strconv.Itoa(tmpl_topic_alt_vars.Topic.CreatedBy) + `" class="the_name">` + tmpl_topic_alt_vars.Topic.CreatedByName + `</a>
			`))
if tmpl_topic_alt_vars.Topic.Tag != "" {
w.Write([]byte(`<div class="tag_block"><div class="tag_pre"></div><div class="post_tag">` + tmpl_topic_alt_vars.Topic.Tag + `</div><div class="tag_post"></div></div>`))
}
w.Write([]byte(`
		</div>
		<div class="content_container">
			<div class="hide_on_edit topic_content user_content nobuttons">` + string(tmpl_topic_alt_vars.Topic.Content.(template.HTML)) + `</div>
			<textarea name="topic_content" class="show_on_edit topic_content_input">` + string(tmpl_topic_alt_vars.Topic.Content.(template.HTML)) + `</textarea>
		</div>
		<div style="clear:both;"></div>
	</div>
	`))
if len(tmpl_topic_alt_vars.ItemList) != 0 {
for _, item := range tmpl_topic_alt_vars.ItemList {
w.Write([]byte(`
	<div class="rowitem passive deletable_block editable_parent post_item">
		<div class="userinfo">
			<div class="avatar_item" style="background-image: url(` + item.Avatar + `), url(/static/white-dot.jpg);background-position: 0px -10px;">&nbsp;</div>
			<a href="/user/` + strconv.Itoa(item.CreatedBy) + `" class="the_name">` + item.CreatedByName + `</a>
			`))
if item.Tag != "" {
w.Write([]byte(`<div class="tag_block"><div class="tag_pre"></div><div class="post_tag">` + item.Tag + `</div><div class="tag_post"></div></div>`))
}
w.Write([]byte(`
		</div>
		<div class="content_container">
			<div class="editable_block user_content">` + string(item.ContentHtml) + `</div>
			<div class="button_container">
				`))
if tmpl_topic_alt_vars.CurrentUser.Perms.EditReply {
w.Write([]byte(`<a href="/reply/edit/submit/` + strconv.Itoa(item.ID) + `" class="action_button edit_item">Edit</a>`))
}
w.Write([]byte(`
				`))
if tmpl_topic_alt_vars.CurrentUser.Perms.DeleteReply {
w.Write([]byte(`<a href="/reply/delete/submit/` + strconv.Itoa(item.ID) + `" class="action_button delete_item">Delete</a>`))
}
w.Write([]byte(`
				<a href="/report/submit/` + strconv.Itoa(item.ID) + `?session=` + tmpl_topic_alt_vars.CurrentUser.Session + `&type=reply" class="action_button report_item">Report</a>
			</div>
		</div>
		<div style="clear:both;"></div>
	</div>
`))
}
}
w.Write([]byte(`</div>
`))
if tmpl_topic_alt_vars.CurrentUser.Perms.CreateReply {
w.Write([]byte(`
<div class="rowblock" style="border-top: none;">
	<form action="/reply/create/" method="post">
		<input name="tid" value='` + strconv.Itoa(tmpl_topic_alt_vars.Topic.ID) + `' type="hidden" />
		<div class="formrow">
			<div class="formitem"><textarea name="reply-content" placeholder="Insert reply here"></textarea></div>
		</div>
		<div class="formrow">
			<div class="formitem"><button name="reply-button" class="formbutton">Create Reply</button></div>
		</div>
	</form>
</div>
`))
}
w.Write([]byte(`
			<!--<link rel="stylesheet" href="https://use.fontawesome.com/8670aa03ca.css">-->
		</div><div style="clear: both;"></div></div></div>
	</body>
</html>`))
}
