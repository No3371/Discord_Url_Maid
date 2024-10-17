package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
)

func CleanMessage(message *gateway.MessageCreateEvent, data *Data, s *state.State) {
	// Ignore bot messages
	if message.Author.Bot {
		return
	}

	stats.TotalMessages++

	notUrlOnly, err := multiUrlsOnlyDetector.MatchString(message.Content)
	if err != nil {
		log.Println("Failed to detect if message is URL only:", err)
	}
    err = nil

	// if notUrlOnly {
    //     notUrlOnly, err = urlOnlyDetector.MatchString(message.Content)
    //     if err != nil {
    //         log.Println("Failed to detect if message is URL only:", err)
    //     }
	// }
    // err = nil

	// Find all URLs in the message
	urlMatch, err := urlExtractor.FindStringMatch(message.Content)
	if err != nil {
		log.Println("Failed to find URLs in message:", err)
		return
	}

	containsRedirect := false
	cleaned := false
	var replyString string

	var urlMap = make(map[string]string)

	// Loop through all matches (URLs)
	for urlMatch != nil {
		processed, is_redirect := CleanUrl(urlMatch.String(), data)
		urlMap[urlMatch.String()] = processed

		if is_redirect {
			containsRedirect = true
		}

		if processed != urlMatch.String() {
			cleaned = true
		}

		// Move to the next match (URL)
		urlMatch, err = urlExtractor.FindNextMatch(urlMatch)
		if err != nil {
			log.Println("Failed to find next URL in message:", err)
		}
        err = nil
	}

	if !cleaned && !containsRedirect {
		return
	}

	sb := strings.Builder{}

	for url, processed := range urlMap {
		if url != processed || cleaned {
			sb.WriteString(processed)
			sb.WriteRune('\n')
		}
	}

	if containsRedirect {
		sb.WriteString("↪️ Redirect Found / 此訊息包含自動轉址\n")
	}

	if cleaned {
		stats.CleanedMessages++
	}

	replyString = sb.String()
	if len(replyString) > 0 && replyString[len(replyString)-1] == '\n' {
		replyString = replyString[:len(replyString)-1]
	}

	edit := api.EditMessageData{}
	edit.Flags = new(discord.MessageFlags)
	*edit.Flags = message.Flags
	*edit.Flags |= discord.SuppressEmbeds
	_, err = s.EditMessageComplex(message.ChannelID, message.ID, edit)
	if err != nil {
		log.Printf("Failed to edit message: %v", err)
		return
	}

	msgData := api.SendMessageData{
		AllowedMentions: allowedMentions,
		Reference: &discord.MessageReference{
			MessageID: message.ID,
			ChannelID: message.ChannelID,
			GuildID:   message.GuildID,
		},
		Flags: discord.SuppressNotifications,
	}

	if notUrlOnly {
		msgData.Content = replyString
		msgData.Reference = nil
	} else {
		if len(urlMap) > 1 {
			msgData.Content = fmt.Sprintf("%s:\n%s", message.Author.Mention(), replyString)
		} else {
			msgData.Content = fmt.Sprintf("%s: %s", message.Author.Mention(), replyString)
		}
	}

	_, err = s.SendMessageComplex(message.ChannelID, msgData)
	if err != nil {
		log.Printf("Failed to reply: %v", err)
	}
    err = nil

	if cleaned && !notUrlOnly {
		err := s.DeleteMessage(message.ChannelID, message.ID, "URL only message")
		if err != nil {
			log.Printf("Failed to delete message: %v", err)
		}
        err = nil
	}
}

func CleanUrl(url string, data *Data) (processed string, is_redirect bool) {
	siteRulesApplied := false
	processed = url

	// Loop through each provider
	for name, provider := range data.Providers {
		// Optimization: Skip site rules if already applied
		if siteRulesApplied && !strings.HasPrefix(name, "globalRules") {
			continue
		}

		if match, _ := provider.UrlPattern.MatchString(processed); !match {
			continue // next provider
		}

		if !strings.HasPrefix(name, "globalRules") {
			siteRulesApplied = true
		}

		for _, rdr := range provider.Redirections {
			if ridrectFound, _ := rdr.MatchString(processed); ridrectFound {
				stats.Redirects++
				is_redirect = true
				continue
			}
		}

		// Skip URL if it matches any exception pattern
		exceptionFound := false
		for _, exception := range provider.Exceptions {
			if exceptionMatch, _ := exception.MatchString(processed); exceptionMatch {
				exceptionFound = true
				break // next exception rule
			}
		}
		if exceptionFound {
			continue // next provider
		}

		paramMatch, err := paramExtracter.FindStringMatch(processed)
		if err != nil {
			log.Println("Failed to find parameters in URL:", err)
			continue
		}

		// Loop through all query parameters
		for paramMatch != nil {
			stats.TotalParams++
			// Extract the param key and value
			paramName := paramMatch.GroupByNumber(1).String()

			// Check if the paramValue matches any of the provider's rules
			for _, rule := range provider.Rules {
				if match, _ := rule.MatchString(paramName); match {
					// Remove or replace based on the initial character ('?' or '&')
					if strings.HasPrefix(paramMatch.String(), "&") {
						processed = strings.Replace(processed, paramMatch.String(), "", 1)
					} else if strings.HasPrefix(paramMatch.String(), "?") {
						processed = strings.Replace(processed, paramMatch.String(), "?", 1)
					}

					stats.CleanedParams++
					break
				}
			}

			paramMatch, err = paramExtracter.FindNextMatch(paramMatch)
			if err != nil {
				log.Println("Failed to find next parameter in URL:", err)
			}
		}
	}

	if processed != url {
		stats.CleanedURLs++
		if processed[len(processed)-1] == '?' {
			processed = processed[:len(processed)-1]
		}
	}

	return processed, is_redirect
}

// cleanTrackingParams removes tracking parameters from any URLs in the message
// CleanMessageAndReport function that processes a message string and cleans up URLs based on providers' rules
// func CleanMessageAndReport(message string, data *Data) string {
// 	stats.TotalMessages++
// 	// Find all URLs in the message
// 	urlMatch, err := urlExtractor.FindStringMatch(message)
// 	if err != nil {
// 		log.Println("Failed to find URLs in message:", err)
// 		return ""
// 	}

// 	containsRedirect := false
// 	modified := false
// 	reply := ""

// 	// Loop through all matches (URLs)
// 	for urlMatch != nil {
// 		stats.TotalURLs++
// 		urlStr := urlMatch.String()
// 		urlModified := false
// 		urlMatched := false

// 		// Loop through each provider
// 		for name, provider := range data.Providers {
// 			if urlMatched && !strings.HasPrefix(name, "globalRules") {
// 				continue
// 			}

// 			if match, _ := provider.UrlPattern.MatchString(urlStr); !match {
// 				continue // next provider
// 			}

// 			if !strings.HasPrefix(name, "globalRules") {
// 				urlMatched = true
// 			}

// 			for _, rdr := range provider.Redirections {
// 				if ridrectFound, _ := rdr.MatchString(urlStr); ridrectFound {
// 					stats.Redirects++
// 					containsRedirect = true
// 					continue
// 				}
// 			}

// 			// Skip URL if it matches any exception pattern
// 			exceptionFound := false
// 			for _, exception := range provider.Exceptions {
// 				if exceptionMatch, _ := exception.MatchString(urlStr); exceptionMatch {
// 					exceptionFound = true
// 					break // next exception rule
// 				}
// 			}
// 			if exceptionFound {
// 				continue // next provider
// 			}

// 			paramMatch, err := paramExtracter.FindStringMatch(urlStr)
// 			if err != nil {
// 				log.Println("Failed to find parameters in URL:", err)
// 				continue
// 			}

// 			// Loop through all query parameters
// 			for paramMatch != nil {
// 				stats.TotalParams++
// 				// Extract the param key and value
// 				paramName := paramMatch.GroupByNumber(1).String()

// 				// Check if the paramValue matches any of the provider's rules
// 				for _, rule := range provider.Rules {
// 					if match, _ := rule.MatchString(paramName); match {
// 						// Remove or replace based on the initial character ('?' or '&')
// 						if strings.HasPrefix(paramMatch.String(), "&") {
// 							urlStr = strings.Replace(urlStr, paramMatch.String(), "", 1)
// 						} else if strings.HasPrefix(paramMatch.String(), "?") {
// 							urlStr = strings.Replace(urlStr, paramMatch.String(), "?", 1)
// 						}

// 						stats.CleanedParams++
// 						modified = true
// 						urlModified = true
// 						break
// 					}
// 				}

// 				paramMatch, err = paramExtracter.FindNextMatch(paramMatch)
// 				if err != nil {
// 					log.Println("Failed to find next parameter in URL:", err)
// 				}
// 			}
// 		}

// 		if urlModified {
// 			stats.CleanedURLs++
// 			if urlStr[len(urlStr)-1] == '?' {
// 				urlStr = urlStr[:len(urlStr)-1]
// 			}
// 			reply += urlStr + "\n"
// 		}

// 		// Move to the next match (URL)
// 		urlMatch, err = urlExtractor.FindNextMatch(urlMatch)
// 		if err != nil {
// 			log.Println("Failed to find next URL in message:", err)
// 		}
// 	}

// 	if containsRedirect {
// 		reply = reply + "↪️ Redirect Found / 此訊息包含自動轉址\n"
// 	}

// 	if modified {
// 		stats.CleanedMessages++
// 	}

// 	if len(reply) > 0 && reply[len(reply)-1] == '\n' {
// 		reply = reply[:len(reply)-1]
// 	}
// 	// Return the cleaned-up message
// 	return reply
// }
