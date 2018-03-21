package main

import (
	"fmt"
	"testing"
)

func TestFormatComment(t *testing.T) {
	text := "Reloads the script that is set in an app.\n\n### Logs\nWhen this method is called, audit activity logs are produced to track the user activity.\nIn the case of errors, both audit activity logs and system services logs are produced.\nThe log files are named as follows:\n<table>\n<tr>\n<th>Audit activity log</th>\n<th>System service log</th>\n</tr>\n<tr>\n<td>&lt;i&gt;&lt; MachineName&gt;&lt;/i&gt;AuditActivity&lt;i&gt;Engine.txt&lt;/i&gt; in Qlik Sense Enterprise&lt;br&gt;&lt;i&gt;&lt; MachineName&gt;&lt;/i&gt;AuditActivity&lt;i&gt;Engine.log&lt;/i&gt; in Qlik Sense Desktop</td>\n<td>&lt;i&gt;&lt; MachineName&gt;&lt;/i&gt;Service&lt;i&gt;Engine.txt&lt;/i&gt; in Qlik Sense Enterprise&lt;br&gt;&lt;i&gt;&lt; MachineName&gt;&lt;/i&gt;Service&lt;i&gt;Engine.log&lt;/i&gt; in Qlik Sense Desktop</td>\n</tr>\n</table>\n\n### Where to find the log files\nThe location of the log files depends on whether you have installed Qlik Sense Enterprise or Qlik Sense Desktop.\n<table>\n<tr>\n<th>Qlik Sense Enterprise </th>\n<th>Qlik Sense Desktop </th>\n</tr>\n<tr>\n<td>&lt;i&gt;%ProgramData%/Qlik/Sense/Log/Engine&lt;/i&gt;</td>\n<td>&lt;i&gt;%UserProfile%/Documents/Qlik/Sense/Log&lt;/i&gt;</td>\n</tr>\n</table>"

	fmt.Println(text)
	fmt.Println("-----------------------------------------")
	fmt.Println(formatComment(" ", text, []*Type{}))
}

func TestFormatComment2(t *testing.T) {
	text := "Returns:\n* The list of tables in an app and the fields inside each table.\n* The list of derived fields.\n* The list of key fields."

	fmt.Println(text)
	fmt.Println("-----------------------------------------")
	fmt.Println(formatComment(" ", text, []*Type{}))
}
