// +build !no_templategen

// Code generated by Gosora. More below:
/* This file was automatically generated by the software. Please don't edit it as your changes may be overwritten at any moment. */
package main
import "io"
import "./common"

var forums_tmpl_phrase_id int

// nolint
func init() {
	common.Template_forums_handle = Template_forums
	common.Ctemplates = append(common.Ctemplates,"forums")
	common.TmplPtrMap["forums"] = &common.Template_forums_handle
	common.TmplPtrMap["o_forums"] = Template_forums
	forums_tmpl_phrase_id = common.RegisterTmplPhraseNames([]string{
		"menu_account_aria",
		"menu_account_tooltip",
		"menu_profile_aria",
		"menu_profile_tooltip",
		"menu_panel_aria",
		"menu_panel_tooltip",
		"menu_logout_aria",
		"menu_logout_tooltip",
		"menu_register_aria",
		"menu_register_tooltip",
		"menu_login_aria",
		"menu_login_tooltip",
		"menu_hamburger_tooltip",
		"forums_head",
		"forums_no_description",
		"forums_none",
		"forums_no_forums",
		"footer_powered_by",
		"footer_made_with_love",
		"footer_theme_selector_aria",
	})
}

// nolint
func Template_forums(tmpl_forums_vars common.ForumsPage, w io.Writer) error {
var phrases = common.GetTmplPhrasesBytes(forums_tmpl_phrase_id)
w.Write(header_frags[0])
w.Write([]byte(tmpl_forums_vars.Title))
w.Write(header_frags[1])
w.Write([]byte(tmpl_forums_vars.Header.Site.Name))
w.Write(header_frags[2])
w.Write([]byte(tmpl_forums_vars.Header.Theme.Name))
w.Write(header_frags[3])
if len(tmpl_forums_vars.Header.Stylesheets) != 0 {
for _, item := range tmpl_forums_vars.Header.Stylesheets {
w.Write(header_frags[4])
w.Write([]byte(item))
w.Write(header_frags[5])
}
}
w.Write(header_frags[6])
if len(tmpl_forums_vars.Header.Scripts) != 0 {
for _, item := range tmpl_forums_vars.Header.Scripts {
w.Write(header_frags[7])
w.Write([]byte(item))
w.Write(header_frags[8])
}
}
w.Write(header_frags[9])
w.Write([]byte(tmpl_forums_vars.CurrentUser.Session))
w.Write(header_frags[10])
w.Write([]byte(tmpl_forums_vars.Header.Site.URL))
w.Write(header_frags[11])
if tmpl_forums_vars.Header.MetaDesc != "" {
w.Write(header_frags[12])
w.Write([]byte(tmpl_forums_vars.Header.MetaDesc))
w.Write(header_frags[13])
}
w.Write(header_frags[14])
if !tmpl_forums_vars.CurrentUser.IsSuperMod {
w.Write(header_frags[15])
}
w.Write(header_frags[16])
w.Write(menu_frags[0])
w.Write([]byte(common.BuildWidget("leftOfNav",tmpl_forums_vars.Header)))
w.Write(menu_frags[1])
w.Write([]byte(tmpl_forums_vars.Header.Site.ShortName))
w.Write(menu_frags[2])
w.Write([]byte(common.BuildWidget("topMenu",tmpl_forums_vars.Header)))
if tmpl_forums_vars.CurrentUser.Loggedin {
w.Write(menu_frags[3])
w.Write(phrases[0])
w.Write(menu_frags[4])
w.Write(phrases[1])
w.Write(menu_frags[5])
w.Write([]byte(tmpl_forums_vars.CurrentUser.Link))
w.Write(menu_frags[6])
w.Write(phrases[2])
w.Write(menu_frags[7])
w.Write(phrases[3])
w.Write(menu_frags[8])
w.Write(phrases[4])
w.Write(menu_frags[9])
w.Write(phrases[5])
w.Write(menu_frags[10])
w.Write([]byte(tmpl_forums_vars.CurrentUser.Session))
w.Write(menu_frags[11])
w.Write(phrases[6])
w.Write(menu_frags[12])
w.Write(phrases[7])
w.Write(menu_frags[13])
} else {
w.Write(menu_frags[14])
w.Write(phrases[8])
w.Write(menu_frags[15])
w.Write(phrases[9])
w.Write(menu_frags[16])
w.Write(phrases[10])
w.Write(menu_frags[17])
w.Write(phrases[11])
w.Write(menu_frags[18])
}
w.Write(menu_frags[19])
w.Write(phrases[12])
w.Write(menu_frags[20])
w.Write([]byte(common.BuildWidget("rightOfNav",tmpl_forums_vars.Header)))
w.Write(menu_frags[21])
w.Write(header_frags[17])
if tmpl_forums_vars.Header.Widgets.RightSidebar != "" {
w.Write(header_frags[18])
}
w.Write(header_frags[19])
if len(tmpl_forums_vars.Header.NoticeList) != 0 {
for _, item := range tmpl_forums_vars.Header.NoticeList {
w.Write(header_frags[20])
w.Write([]byte(item))
w.Write(header_frags[21])
}
}
w.Write(header_frags[22])
w.Write(forums_frags[0])
w.Write(phrases[13])
w.Write(forums_frags[1])
if len(tmpl_forums_vars.ItemList) != 0 {
for _, item := range tmpl_forums_vars.ItemList {
w.Write(forums_frags[2])
if item.Desc != "" || item.LastTopic.Title != "" {
w.Write(forums_frags[3])
}
w.Write(forums_frags[4])
w.Write([]byte(item.Link))
w.Write(forums_frags[5])
w.Write([]byte(item.Name))
w.Write(forums_frags[6])
if item.Desc != "" {
w.Write(forums_frags[7])
w.Write([]byte(item.Desc))
w.Write(forums_frags[8])
} else {
w.Write(forums_frags[9])
w.Write(phrases[14])
w.Write(forums_frags[10])
}
w.Write(forums_frags[11])
if item.LastReplyer.Avatar != "" {
w.Write(forums_frags[12])
w.Write([]byte(item.LastReplyer.Avatar))
w.Write(forums_frags[13])
w.Write([]byte(item.LastReplyer.Name))
w.Write(forums_frags[14])
w.Write([]byte(item.LastReplyer.Name))
w.Write(forums_frags[15])
}
w.Write(forums_frags[16])
w.Write([]byte(item.LastTopic.Link))
w.Write(forums_frags[17])
if item.LastTopic.Title != "" {
w.Write([]byte(item.LastTopic.Title))
} else {
w.Write(phrases[15])
}
w.Write(forums_frags[18])
if item.LastTopicTime != "" {
w.Write(forums_frags[19])
w.Write([]byte(item.LastTopicTime))
w.Write(forums_frags[20])
}
w.Write(forums_frags[21])
}
} else {
w.Write(forums_frags[22])
w.Write(phrases[16])
w.Write(forums_frags[23])
}
w.Write(forums_frags[24])
w.Write(footer_frags[0])
w.Write([]byte(common.BuildWidget("footer",tmpl_forums_vars.Header)))
w.Write(footer_frags[1])
w.Write(phrases[17])
w.Write(footer_frags[2])
w.Write(phrases[18])
w.Write(footer_frags[3])
w.Write(phrases[19])
w.Write(footer_frags[4])
if len(tmpl_forums_vars.Header.Themes) != 0 {
for _, item := range tmpl_forums_vars.Header.Themes {
if !item.HideFromThemes {
w.Write(footer_frags[5])
w.Write([]byte(item.Name))
w.Write(footer_frags[6])
if tmpl_forums_vars.Header.Theme.Name == item.Name {
w.Write(footer_frags[7])
}
w.Write(footer_frags[8])
w.Write([]byte(item.FriendlyName))
w.Write(footer_frags[9])
}
}
}
w.Write(footer_frags[10])
w.Write([]byte(common.BuildWidget("rightSidebar",tmpl_forums_vars.Header)))
w.Write(footer_frags[11])
	return nil
}
