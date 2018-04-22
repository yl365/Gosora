// +build !pgsql, !sqlite, !mssql

/* This file was generated by Gosora's Query Generator. Please try to avoid modifying this file, as it might change at any time. */

package main

import "log"
import "database/sql"
import "./common"
//import "./query_gen/lib"

// nolint
type Stmts struct {
	getPassword *sql.Stmt
	isPluginActive *sql.Stmt
	getUsersOffset *sql.Stmt
	isThemeDefault *sql.Stmt
	getModlogs *sql.Stmt
	getModlogsOffset *sql.Stmt
	getAdminlogsOffset *sql.Stmt
	getTopicFID *sql.Stmt
	getUserName *sql.Stmt
	getEmailsByUser *sql.Stmt
	getTopicBasic *sql.Stmt
	forumEntryExists *sql.Stmt
	groupEntryExists *sql.Stmt
	getAttachment *sql.Stmt
	getForumTopics *sql.Stmt
	createReport *sql.Stmt
	addForumPermsToForum *sql.Stmt
	addPlugin *sql.Stmt
	addTheme *sql.Stmt
	createWordFilter *sql.Stmt
	editReply *sql.Stmt
	updatePlugin *sql.Stmt
	updatePluginInstall *sql.Stmt
	updateTheme *sql.Stmt
	updateUser *sql.Stmt
	updateGroupPerms *sql.Stmt
	updateGroup *sql.Stmt
	updateEmail *sql.Stmt
	verifyEmail *sql.Stmt
	setTempGroup *sql.Stmt
	updateWordFilter *sql.Stmt
	bumpSync *sql.Stmt
	deleteActivityStreamMatch *sql.Stmt
	deleteWordFilter *sql.Stmt
	reportExists *sql.Stmt

	getActivityFeedByWatcher *sql.Stmt
	getActivityCountByWatcher *sql.Stmt
	todaysPostCount *sql.Stmt
	todaysTopicCount *sql.Stmt
	todaysReportCount *sql.Stmt
	todaysNewUserCount *sql.Stmt

	Mocks bool
}

// nolint
func _gen_mysql() (err error) {
	common.DebugLog("Building the generated statements")
	
	common.DebugLog("Preparing getPassword statement.")
	stmts.getPassword, err = db.Prepare("SELECT `password`,`salt` FROM `users` WHERE `uid` = ?")
	if err != nil {
		log.Print("Error in getPassword statement.")
		return err
	}
		
	common.DebugLog("Preparing isPluginActive statement.")
	stmts.isPluginActive, err = db.Prepare("SELECT `active` FROM `plugins` WHERE `uname` = ?")
	if err != nil {
		log.Print("Error in isPluginActive statement.")
		return err
	}
		
	common.DebugLog("Preparing getUsersOffset statement.")
	stmts.getUsersOffset, err = db.Prepare("SELECT `uid`,`name`,`group`,`active`,`is_super_admin`,`avatar` FROM `users` ORDER BY `uid` ASC LIMIT ?,?")
	if err != nil {
		log.Print("Error in getUsersOffset statement.")
		return err
	}
		
	common.DebugLog("Preparing isThemeDefault statement.")
	stmts.isThemeDefault, err = db.Prepare("SELECT `default` FROM `themes` WHERE `uname` = ?")
	if err != nil {
		log.Print("Error in isThemeDefault statement.")
		return err
	}
		
	common.DebugLog("Preparing getModlogs statement.")
	stmts.getModlogs, err = db.Prepare("SELECT `action`,`elementID`,`elementType`,`ipaddress`,`actorID`,`doneAt` FROM `moderation_logs`")
	if err != nil {
		log.Print("Error in getModlogs statement.")
		return err
	}
		
	common.DebugLog("Preparing getModlogsOffset statement.")
	stmts.getModlogsOffset, err = db.Prepare("SELECT `action`,`elementID`,`elementType`,`ipaddress`,`actorID`,`doneAt` FROM `moderation_logs` ORDER BY `doneAt` DESC LIMIT ?,?")
	if err != nil {
		log.Print("Error in getModlogsOffset statement.")
		return err
	}
		
	common.DebugLog("Preparing getAdminlogsOffset statement.")
	stmts.getAdminlogsOffset, err = db.Prepare("SELECT `action`,`elementID`,`elementType`,`ipaddress`,`actorID`,`doneAt` FROM `administration_logs` ORDER BY `doneAt` DESC LIMIT ?,?")
	if err != nil {
		log.Print("Error in getAdminlogsOffset statement.")
		return err
	}
		
	common.DebugLog("Preparing getTopicFID statement.")
	stmts.getTopicFID, err = db.Prepare("SELECT `parentID` FROM `topics` WHERE `tid` = ?")
	if err != nil {
		log.Print("Error in getTopicFID statement.")
		return err
	}
		
	common.DebugLog("Preparing getUserName statement.")
	stmts.getUserName, err = db.Prepare("SELECT `name` FROM `users` WHERE `uid` = ?")
	if err != nil {
		log.Print("Error in getUserName statement.")
		return err
	}
		
	common.DebugLog("Preparing getEmailsByUser statement.")
	stmts.getEmailsByUser, err = db.Prepare("SELECT `email`,`validated`,`token` FROM `emails` WHERE `uid` = ?")
	if err != nil {
		log.Print("Error in getEmailsByUser statement.")
		return err
	}
		
	common.DebugLog("Preparing getTopicBasic statement.")
	stmts.getTopicBasic, err = db.Prepare("SELECT `title`,`content` FROM `topics` WHERE `tid` = ?")
	if err != nil {
		log.Print("Error in getTopicBasic statement.")
		return err
	}
		
	common.DebugLog("Preparing forumEntryExists statement.")
	stmts.forumEntryExists, err = db.Prepare("SELECT `fid` FROM `forums` WHERE `name` = '' ORDER BY `fid` ASC LIMIT 0,1")
	if err != nil {
		log.Print("Error in forumEntryExists statement.")
		return err
	}
		
	common.DebugLog("Preparing groupEntryExists statement.")
	stmts.groupEntryExists, err = db.Prepare("SELECT `gid` FROM `users_groups` WHERE `name` = '' ORDER BY `gid` ASC LIMIT 0,1")
	if err != nil {
		log.Print("Error in groupEntryExists statement.")
		return err
	}
		
	common.DebugLog("Preparing getAttachment statement.")
	stmts.getAttachment, err = db.Prepare("SELECT `sectionID`,`sectionTable`,`originID`,`originTable`,`uploadedBy`,`path` FROM `attachments` WHERE `path` = ? AND `sectionID` = ? AND `sectionTable` = ?")
	if err != nil {
		log.Print("Error in getAttachment statement.")
		return err
	}
		
	common.DebugLog("Preparing getForumTopics statement.")
	stmts.getForumTopics, err = db.Prepare("SELECT `topics`.`tid`, `topics`.`title`, `topics`.`content`, `topics`.`createdBy`, `topics`.`is_closed`, `topics`.`sticky`, `topics`.`createdAt`, `topics`.`lastReplyAt`, `topics`.`parentID`, `users`.`name`, `users`.`avatar` FROM `topics` LEFT JOIN `users` ON `topics`.`createdBy` = `users`.`uid`  WHERE `topics`.`parentID` = ? ORDER BY `topics`.`sticky` DESC,`topics`.`lastReplyAt` DESC,`topics`.`createdBy` DESC")
	if err != nil {
		log.Print("Error in getForumTopics statement.")
		return err
	}
		
	common.DebugLog("Preparing createReport statement.")
	stmts.createReport, err = db.Prepare("INSERT INTO `topics`(`title`,`content`,`parsed_content`,`createdAt`,`lastReplyAt`,`createdBy`,`lastReplyBy`,`data`,`parentID`,`css_class`) VALUES (?,?,?,UTC_TIMESTAMP(),UTC_TIMESTAMP(),?,?,?,1,'report')")
	if err != nil {
		log.Print("Error in createReport statement.")
		return err
	}
		
	common.DebugLog("Preparing addForumPermsToForum statement.")
	stmts.addForumPermsToForum, err = db.Prepare("INSERT INTO `forums_permissions`(`gid`,`fid`,`preset`,`permissions`) VALUES (?,?,?,?)")
	if err != nil {
		log.Print("Error in addForumPermsToForum statement.")
		return err
	}
		
	common.DebugLog("Preparing addPlugin statement.")
	stmts.addPlugin, err = db.Prepare("INSERT INTO `plugins`(`uname`,`active`,`installed`) VALUES (?,?,?)")
	if err != nil {
		log.Print("Error in addPlugin statement.")
		return err
	}
		
	common.DebugLog("Preparing addTheme statement.")
	stmts.addTheme, err = db.Prepare("INSERT INTO `themes`(`uname`,`default`) VALUES (?,?)")
	if err != nil {
		log.Print("Error in addTheme statement.")
		return err
	}
		
	common.DebugLog("Preparing createWordFilter statement.")
	stmts.createWordFilter, err = db.Prepare("INSERT INTO `word_filters`(`find`,`replacement`) VALUES (?,?)")
	if err != nil {
		log.Print("Error in createWordFilter statement.")
		return err
	}
		
	common.DebugLog("Preparing editReply statement.")
	stmts.editReply, err = db.Prepare("UPDATE `replies` SET `content` = ?,`parsed_content` = ? WHERE `rid` = ?")
	if err != nil {
		log.Print("Error in editReply statement.")
		return err
	}
		
	common.DebugLog("Preparing updatePlugin statement.")
	stmts.updatePlugin, err = db.Prepare("UPDATE `plugins` SET `active` = ? WHERE `uname` = ?")
	if err != nil {
		log.Print("Error in updatePlugin statement.")
		return err
	}
		
	common.DebugLog("Preparing updatePluginInstall statement.")
	stmts.updatePluginInstall, err = db.Prepare("UPDATE `plugins` SET `installed` = ? WHERE `uname` = ?")
	if err != nil {
		log.Print("Error in updatePluginInstall statement.")
		return err
	}
		
	common.DebugLog("Preparing updateTheme statement.")
	stmts.updateTheme, err = db.Prepare("UPDATE `themes` SET `default` = ? WHERE `uname` = ?")
	if err != nil {
		log.Print("Error in updateTheme statement.")
		return err
	}
		
	common.DebugLog("Preparing updateUser statement.")
	stmts.updateUser, err = db.Prepare("UPDATE `users` SET `name` = ?,`email` = ?,`group` = ? WHERE `uid` = ?")
	if err != nil {
		log.Print("Error in updateUser statement.")
		return err
	}
		
	common.DebugLog("Preparing updateGroupPerms statement.")
	stmts.updateGroupPerms, err = db.Prepare("UPDATE `users_groups` SET `permissions` = ? WHERE `gid` = ?")
	if err != nil {
		log.Print("Error in updateGroupPerms statement.")
		return err
	}
		
	common.DebugLog("Preparing updateGroup statement.")
	stmts.updateGroup, err = db.Prepare("UPDATE `users_groups` SET `name` = ?,`tag` = ? WHERE `gid` = ?")
	if err != nil {
		log.Print("Error in updateGroup statement.")
		return err
	}
		
	common.DebugLog("Preparing updateEmail statement.")
	stmts.updateEmail, err = db.Prepare("UPDATE `emails` SET `email` = ?,`uid` = ?,`validated` = ?,`token` = ? WHERE `email` = ?")
	if err != nil {
		log.Print("Error in updateEmail statement.")
		return err
	}
		
	common.DebugLog("Preparing verifyEmail statement.")
	stmts.verifyEmail, err = db.Prepare("UPDATE `emails` SET `validated` = 1,`token` = '' WHERE `email` = ?")
	if err != nil {
		log.Print("Error in verifyEmail statement.")
		return err
	}
		
	common.DebugLog("Preparing setTempGroup statement.")
	stmts.setTempGroup, err = db.Prepare("UPDATE `users` SET `temp_group` = ? WHERE `uid` = ?")
	if err != nil {
		log.Print("Error in setTempGroup statement.")
		return err
	}
		
	common.DebugLog("Preparing updateWordFilter statement.")
	stmts.updateWordFilter, err = db.Prepare("UPDATE `word_filters` SET `find` = ?,`replacement` = ? WHERE `wfid` = ?")
	if err != nil {
		log.Print("Error in updateWordFilter statement.")
		return err
	}
		
	common.DebugLog("Preparing bumpSync statement.")
	stmts.bumpSync, err = db.Prepare("UPDATE `sync` SET `last_update` = UTC_TIMESTAMP()")
	if err != nil {
		log.Print("Error in bumpSync statement.")
		return err
	}
		
	common.DebugLog("Preparing deleteActivityStreamMatch statement.")
	stmts.deleteActivityStreamMatch, err = db.Prepare("DELETE FROM `activity_stream_matches` WHERE `watcher` = ? AND `asid` = ?")
	if err != nil {
		log.Print("Error in deleteActivityStreamMatch statement.")
		return err
	}
		
	common.DebugLog("Preparing deleteWordFilter statement.")
	stmts.deleteWordFilter, err = db.Prepare("DELETE FROM `word_filters` WHERE `wfid` = ?")
	if err != nil {
		log.Print("Error in deleteWordFilter statement.")
		return err
	}
		
	common.DebugLog("Preparing reportExists statement.")
	stmts.reportExists, err = db.Prepare("SELECT COUNT(*) AS `count` FROM `topics` WHERE `data` = ? AND `data` != '' AND `parentID` = 1")
	if err != nil {
		log.Print("Error in reportExists statement.")
		return err
	}
	
	return nil
}
