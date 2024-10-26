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


type processedUrl struct {
	Url       string
	IsSpoiler bool
	IsRedirect bool
}

func TryCleanMessage(message *gateway.MessageCreateEvent, data *Data, s *state.State) {
	// Ignore bot messages
	if message.Author.Bot {
		return
	}

	stats.TotalMessages++

	urlMap, cleaned, containsRedirect, notUrlOnly, err := TryCleanString(message.Content, data)
	if err != nil {
		log.Println("Failed to clean message:", err)
		return
	}

	if !cleaned && !containsRedirect {
		return
	}

	replyString := PrepareReply(urlMap, containsRedirect, cleaned)

	msgData := api.SendMessageData{
		AllowedMentions: allowedMentions,
		Reference: &discord.MessageReference{
			MessageID: message.ID,
			ChannelID: message.ChannelID,
			GuildID:   message.GuildID,
		},
		Flags: discord.SuppressNotifications,
		// Components: discord.Components (
		// 	&discord.ButtonComponent{
		// 		Label: "♻️",
		// 		Style: discord.SecondaryButtonStyle,
		// 		CustomID: "clean_message",
		// 	},
		// ),
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

	if cleaned && !notUrlOnly && !containsRedirect {
		err := s.DeleteMessage(message.ChannelID, message.ID, "URL only message")
		if err != nil {
			log.Printf("Failed to delete message: %v", err)
		}
		err = nil
		return
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
}

func PrepareReply(urlMap map[string]processedUrl, containsRedirect bool, cleaned bool) string {
	sb := strings.Builder{}

	for _, processed := range urlMap {
		if processed.IsRedirect {
			continue
		}
		if processed.IsSpoiler {
			sb.WriteString("||")
		}
		sb.WriteString(processed.Url)
		if processed.IsSpoiler {
			sb.WriteString("||")
		}
		sb.WriteRune('\n')
	}
	log.Printf("---\n")

	if containsRedirect {
		sb.WriteString("↪️ Redirect Found / 此訊息包含自動轉址（將經由該站點轉向未知站點）\n")
	}

	if cleaned {
		stats.CleanedMessages++
	}

	replyString := sb.String()
	if len(replyString) > 0 && replyString[len(replyString)-1] == '\n' {
		replyString = replyString[:len(replyString)-1]
	}
	return replyString
}

func TryCleanString(str string, data *Data) (urlMap map[string]processedUrl, cleaned bool, containsRedirect bool, notUrlOnly bool, err error) {

	str, err = connectedUrlFinder.Replace(str, "$& ", -1, -1)
	if err != nil {
		log.Println("Failed to fix connected URLs:", err)
		return
	}

	deSpoiled, err := spoilerFinder.Replace(str, " $1 ", -1, -1)
	if err != nil {
		log.Println("Failed to despoil message:", err)
		return
	}
	notUrlOnly, err = impureUrlsDetector.MatchString(deSpoiled)
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

	messageContent := str
	messageContent, err = enforceSpoilerPadding(messageContent)
	if err != nil {
		log.Println("Failed to ensure spoiler edge:", err)
		messageContent = str // failsafe
	}

	// Find all URLs in the message
	urlMatch, err := urlExtractor.FindStringMatch(messageContent)
	if err != nil {
		log.Println("Failed to find URLs in message:", err)
		return
	}

	var cleanedLookup map[string]string

	// Loop through all matches (URLs)
	for urlMatch != nil {

		matched := urlMatch.String()
		url := matched

		processed, is_redirect := CleanUrl(url, data)

		if cleanedLookup == nil {
			cleanedLookup = make(map[string]string)
		}
		if _, ok := cleanedLookup[processed]; !ok {
			if is_redirect {
				containsRedirect = true
				log.Printf("\nFound Redirect: %s", url)
			}
	
			if processed != urlMatch.String() {
				cleaned = true
				log.Printf("\nCleaned: %s -> %s", url, processed)
				cleanedLookup[processed] = url
			}
			if urlMap == nil {
				urlMap = make(map[string]processedUrl)
			}
			urlMap[url] = processedUrl{Url: processed, IsSpoiler: false, IsRedirect: is_redirect}
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

	// Check for urls in spoilers
	// Loop through all spoiler blocks and check if the processed urls are contained
	spoilerMatch, err := spoilerFinder.FindStringMatch(messageContent)
	if err != nil {
		log.Println("Failed to find spoilers in message:", err)
		return
	}
	for spoilerMatch != nil {
		spoiler := spoilerMatch.String()

		for k, url := range urlMap {
			if strings.Contains(spoiler, url.Url) {
				urlMap[k] = processedUrl{Url: url.Url, IsSpoiler: true}
			}
		}

		spoilerMatch, err = spoilerFinder.FindNextMatch(spoilerMatch)
		if err != nil {
			log.Println("Failed to find spoilers in message:", err)
			break
		}
	}

	return
}

func CleanUrl(url string, data *Data) (processed string, is_redirect bool) {

	processed = url

	// Loop through each provider
	for _, provider := range data.Providers {
		processed, is_redirect = applyRules(provider, processed, is_redirect)
		if processed != url {
			break
		}
	}

	// Always apply global rules
	processed, is_redirect = applyRules(data.GlobalRules, processed, is_redirect)

	if processed != url {
		stats.CleanedURLs++
		if len(processed) > 0 && processed[len(processed)-1] == '?' {
			processed = processed[:len(processed)-1]
		}
	}

	return processed, is_redirect
}

func applyRules(provider Provider, url string, is_redirect bool) (string, bool) {

	if match, _ := provider.UrlPattern.MatchString(url); !match {
		return url, is_redirect
	}

	for _, rdr := range provider.Redirections {
		if ridrectFound, _ := rdr.MatchString(url); ridrectFound {
			stats.Redirects++
			is_redirect = true
			continue
		}
	}

	exceptionFound := false
	for _, exception := range provider.Exceptions {
		if exceptionMatch, _ := exception.MatchString(url); exceptionMatch {
			exceptionFound = true
			break
		}
	}
	if exceptionFound {
		return url, is_redirect
	}

	paramMatch, err := paramExtracter.FindStringMatch(url)
	if err != nil {
		log.Println("Failed to find parameters in URL:", err)
		return url, is_redirect
	}

	for paramMatch != nil {
		stats.TotalParams++
		var matchedParam string = paramMatch.String()
		paramName := paramMatch.GroupByNumber(1).String()

		for _, rule := range provider.Rules {
			if match, _ := rule.MatchString(paramName); match {

				if strings.HasPrefix(matchedParam, "&") {
					url = strings.Replace(url, matchedParam, "", 1)
				} else if strings.HasPrefix(matchedParam, "?") {
					url = strings.Replace(url, matchedParam, "?", 1)
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
	return url, is_redirect
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
