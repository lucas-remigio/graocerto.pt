// Package prompts holds the LLM prompt templates as embedded files, so a
// running binary never depends on its working directory to find them.
package prompts

import _ "embed"

//go:embed telegramTransactions.txt
var TelegramTransactions string
