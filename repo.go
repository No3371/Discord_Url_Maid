package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dlclark/regexp2"
)

const repo string = "https://rules2.clearurls.xyz/data.minify.json"

var spoilerFinder = regexp2.MustCompile(`\|\|(\s*?[\s\S]+?\s*)\|\|`, regexp2.None)

func enforceSpoilerPadding(src string) (string, error) {
	return spoilerFinder.Replace(src, "|| $1 ||", -1, -1)
}

var connectedUrlFinder = regexp2.MustCompile(`https?:\/\/\S+?(?=https?:\/\/)`, regexp2.None)

// var linebreaksFinder = regexp2.MustCompile(`\r?\n|\r`, regexp2.None)
var maskedLinkFinder = regexp2.MustCompile(`\[((?!\s*\])[\s\S]+?)\]\([\s　]*(<)?(https?:\/\/(?(2)[^\s>]+|\S+))(?(2)>?)[\s　]*\)`, regexp2.None)

func enforceMaskedLinkPadding(src string) (string, error) {
	return maskedLinkFinder.Replace(src, "[$1]( $3 )", -1, -1)
}

var dcMaskFilter = regexp2.MustCompile(`https?:\/\/\S\S`, regexp2.None)

// var spoilerExtractor = regexp2.MustCompile(`\|\|(\s*?[\s\S]+?\s*)\|\|`, regexp2.None)
// var spoilerExtractor = regexp2.MustCompile(`\|\|\s*(.+?)\s*\|\|`, regexp2.None)

var impureUrlsDetector = regexp2.MustCompile(`^(?!\s*(?:https?:\/\/\S+\.\S+\s*)+$).+`, regexp2.Multiline)

// var impureUrlsDetector = regexp2.MustCompile(`^(?!\s*(?:(?:\s*\|\|)?\s*https?:\/\/\S+\.\S+\s*(?:\|\|\s*)?)+$).+`, regexp2.Multiline) // This version handles discord spoiler syntax ||
// var urlOnlyDetector = regexp2.MustCompile(`^[^\S\r\n]*https?:\/\/\S+$`, regexp2.None)
var urlExtractor = regexp2.MustCompile(`https?:\/\/\S+\.\S+`, regexp2.None)

// var urlExtractor = regexp2.MustCompile(`(?:\|\|\s*)https?:\/\/\S+?\.[^\s|]+(?:\s*\|\|)|https?:\/\/\S+?\.[^\s|]+`, regexp2.None) // [^\s|]+ for Discord
var paramExtracter = regexp2.MustCompile(`[?&]([\w]+)=([\w-\.\*=]+)`, regexp2.None)

// var spoilerExtractor = regexp2.MustCompile(`(?<=\|\|\s*)https?:\/\/\S+(?=\s*\|\|)`, regexp2.None) // \|\|\s*(https?:\/\/\S+?)\s*\|\|

// func Despoil(src string) string {
// 	spoilerMatch, err := spoilerExtractor.FindStringMatch(src)
// 	if err != nil {
// 		return src
// 	}
// 	if spoilerMatch != nil {
// 		return spoilerMatch.String()
// 	}
// 	return src
// }

// Provider represents a single provider from the ClearURLs data
type Provider struct {
	UrlPattern        *regexp2.Regexp   `json:"-"`
	Rules             []*regexp2.Regexp `json:"-"`
	Exceptions        []*regexp2.Regexp `json:"-"`
	IgnoredParameters []*regexp2.Regexp `json:"-"`
	Redirections      []*regexp2.Regexp `json:"-"`
	Aliases           []*regexp2.Regexp `json:"-"`
}

// rawProvider is used for intermediate JSON unmarshalling to keep the string values temporarily
type rawProvider struct {
	UrlPatternStr        string   `json:"urlPattern"`
	RulesStr             []string `json:"rules"`
	ExceptionsStr        []string `json:"exceptions"`
	IgnoredParametersStr []string `json:"ignoredParameters"`
	RedirectionsStr      []string `json:"redirections"`
}

// Data represents the full JSON structure with all providers
type Data struct {
	GlobalRules Provider            `json:"-"`
	Providers   map[string]Provider `json:"-"`
}

const ONLINE_RULES_FILE = "clear_urls_rules.json"
const CUSTOM_RULES_FILE = "custom_rules.json"
const ALIAS_FILE = "aliases.json"

// rawAlias is used for intermediate JSON unmarshalling to keep the string values temporarily
type rawAlias struct {
	UrlPatternStr  string `json:"urlPattern"`
	TargetRuleName string `json:"targetRuleName"`
}

// FetchAndLoadRules fetches the JSON file from the given URL and unmarshals it into the given Data struct
func FetchAndLoadRules(url string) (*Data, error) {
	var raw string
	var fetch bool

	f, err := os.Open(ONLINE_RULES_FILE)
	if err != nil {
		fetch = true
	} else {
		defer f.Close()
		fi, err := f.Stat()
		if err != nil || fi != nil && time.Since(fi.ModTime()) > time.Hour*6 {
			fetch = true
		} else {
			rawBytes, err := io.ReadAll(f)
			if err != nil {
				return nil, fmt.Errorf("readAll: %w", err)
			}
			raw = string(rawBytes)
		}
	}
	if fetch {
		resp, err := http.Get(url)
		if err != nil {
			return nil, fmt.Errorf("get: %w", err)
		}
		defer resp.Body.Close()

		rawBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("readAll: %w", err)
		}
		raw = string(rawBytes)
		f.Close()
		err = os.WriteFile(ONLINE_RULES_FILE, []byte(raw), 0644)
		if err != nil {
			return nil, fmt.Errorf("writeFile: %w", err)
		}

		log.Printf("Updated ClearUrls file cache.")
	}

	// Intermediate structure to hold raw strings
	var rawRepo struct {
		Providers map[string]rawProvider `json:"providers"`
	}
	err = json.NewDecoder(strings.NewReader(raw)).Decode(&rawRepo)
	if err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	var rawData map[string]rawProvider = rawRepo.Providers

	// Initialize final data structure
	data := Data{
		Providers: make(map[string]Provider),
	}

	// Convert rawProvider into Provider with compiled regexes
	for key, rawProvider := range rawData {
		provider, err := makeProvider(key, rawProvider)
		if err != nil {
			return nil, fmt.Errorf("failed to make provider %s: %v", key, err)
		}
		// Add compiled provider to the map
		data.Providers[key] = provider
	}

	data.GlobalRules = data.Providers["globalRules"]
	delete(data.Providers, "globalRules")

	f, err = os.Open(CUSTOM_RULES_FILE)
	if err != nil {
		return &data, nil
	} else {
		defer f.Close()
		rawBytes, err := io.ReadAll(f)
		if err != nil {
			return nil, fmt.Errorf("readAll: %w", err)
		}
		raw = string(rawBytes)
	}

	err = json.NewDecoder(strings.NewReader(raw)).Decode(&rawRepo)
	if err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	rawData = rawRepo.Providers

	// Convert rawProvider into Provider with compiled regexes
	for key, rawProvider := range rawData {
		provider, err := makeProvider(key, rawProvider)
		if err != nil {
			return nil, fmt.Errorf("failed to make provider %s: %v", key, err)
		}

		if key == "globalRules" && provider.UrlPattern.String() == data.GlobalRules.UrlPattern.String() {
			data.GlobalRules.Rules = append(data.GlobalRules.Rules, provider.Rules...)
			data.GlobalRules.Exceptions = append(data.GlobalRules.Exceptions, provider.Exceptions...)
			data.GlobalRules.IgnoredParameters = append(data.GlobalRules.IgnoredParameters, provider.IgnoredParameters...)
			data.GlobalRules.Redirections = append(data.GlobalRules.Redirections, provider.Redirections...)
		} else {
			if existing, ok := data.Providers[key]; ok {
				existing.Rules = append(existing.Rules, provider.Rules...)
				existing.Exceptions = append(existing.Exceptions, provider.Exceptions...)
				existing.IgnoredParameters = append(existing.IgnoredParameters, provider.IgnoredParameters...)
				existing.Redirections = append(existing.Redirections, provider.Redirections...)
				data.Providers[key] = existing
			} else {
				// Add compiled provider to the map
				data.Providers[key] = provider
			}
		}
	}

	f, err = os.Open(ALIAS_FILE)
	if err != nil {
		return &data, nil
	} else {
		defer f.Close()
		rawBytes, err := io.ReadAll(f)
		if err != nil {
			return nil, fmt.Errorf("readAll: %w", err)
		}
		raw = string(rawBytes)
	}

	var rawAliasFile struct {
		Items map[string]rawAlias `json:"aliases"`
	}
	err = json.NewDecoder(strings.NewReader(raw)).Decode(&rawAliasFile)
	if err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	for key, rawAlias := range rawAliasFile.Items {
		target, targetExists := data.Providers[rawAlias.TargetRuleName]

		if !targetExists {
			continue
		}

		urlPattern, err := regexp2.Compile(rawAlias.UrlPatternStr, regexp2.None)
		if err != nil {
			return &data, fmt.Errorf("failed to compile UrlPattern for provider %s: %v", key, err)
		}
		target.Aliases = append(target.Aliases, urlPattern)
		data.Providers[key] = target
	}

	return &data, nil
}

func makeProvider(key string, rawProvider rawProvider) (provider Provider, err error) {
	provider = Provider{}

	// Compile urlPattern
	provider.UrlPattern, err = regexp2.Compile(rawProvider.UrlPatternStr, regexp2.None)
	if err != nil {
		return provider, fmt.Errorf("failed to compile UrlPattern for provider %s: %v", key, err)
	}

	// Compile rules
	for _, ruleStr := range rawProvider.RulesStr {
		rule, err := regexp2.Compile(ruleStr, regexp2.None)
		if err != nil {
			return provider, fmt.Errorf("failed to compile rule for provider %s: %v", key, err)
		}
		provider.Rules = append(provider.Rules, rule)
	}

	// Compile exceptions
	for _, exceptionStr := range rawProvider.ExceptionsStr {
		exception, err := regexp2.Compile(exceptionStr, regexp2.None)
		if err != nil {
			return provider, fmt.Errorf("failed to compile exception for provider %s: %v", key, err)
		}
		provider.Exceptions = append(provider.Exceptions, exception)
	}

	// Compile ignored parameters
	for _, ignoredParamStr := range rawProvider.IgnoredParametersStr {
		ignoredParam, err := regexp2.Compile(ignoredParamStr, regexp2.None)
		if err != nil {
			return provider, fmt.Errorf("failed to compile ignored parameter for provider %s: %v", key, err)
		}
		provider.IgnoredParameters = append(provider.IgnoredParameters, ignoredParam)
	}

	// Compile redirections
	for _, redirectionStr := range rawProvider.RedirectionsStr {
		redirection, err := regexp2.Compile(redirectionStr, regexp2.None)
		if err != nil {
			return provider, fmt.Errorf("failed to compile redirection for provider %s: %v", key, err)
		}
		provider.Redirections = append(provider.Redirections, redirection)
	}

	return provider, nil
}

// let d = [{
// 	character: "h",
// 	matcher: f(["H", "һ", "հ", "Ꮒ", "ℎ", "\uD835\uDC21", "\uD835\uDC89", "\uD835\uDCBD", "\uD835\uDCF1", "\uD835\uDD25", "\uD835\uDD59", "\uD835\uDD8D", "\uD835\uDDC1", "\uD835\uDDF5", "\uD835\uDE29", "\uD835\uDE5D", "\uD835\uDE91", "ｈ"])
// }, {
// 	character: "t",
// 	matcher: f(["T", "\uD835\uDC2D", "\uD835\uDC61", "\uD835\uDC95", "\uD835\uDCC9", "\uD835\uDCFD", "\uD835\uDD31", "\uD835\uDD65", "\uD835\uDD99", "\uD835\uDDCD", "\uD835\uDE01", "\uD835\uDE35", "\uD835\uDE69", "\uD835\uDE9D"])
// }, {
// 	character: "p",
// 	matcher: f(["P", "ρ", "ϱ", "р", "⍴", "ⲣ", "\uD835\uDC29", "\uD835\uDC5D", "\uD835\uDC91", "\uD835\uDCC5", "\uD835\uDCF9", "\uD835\uDD2D", "\uD835\uDD61", "\uD835\uDD95", "\uD835\uDDC9", "\uD835\uDDFD", "\uD835\uDE31", "\uD835\uDE65", "\uD835\uDE99", "\uD835\uDED2", "\uD835\uDEE0", "\uD835\uDF0C", "\uD835\uDF1A", "\uD835\uDF46", "\uD835\uDF54", "\uD835\uDF80", "\uD835\uDF8E", "\uD835\uDFBA", "\uD835\uDFC8", "ｐ", "ҏ"])
// }, {
// 	character: "s",
// 	matcher: f(["S", "ƽ", "ѕ", "ꜱ", "ꮪ", "\uD801\uDC48", "\uD806\uDCC1", "\uD835\uDC2C", "\uD835\uDC60", "\uD835\uDC94", "\uD835\uDCC8", "\uD835\uDCFC", "\uD835\uDD30", "\uD835\uDD64", "\uD835\uDD98", "\uD835\uDDCC", "\uD835\uDE00", "\uD835\uDE34", "\uD835\uDE68", "\uD835\uDE9C", "ｓ"])
// }, {
// 	character: ":",
// 	matcher: f(["ː", "˸", "։", "׃", "܃", "܄", "ः", "ઃ", "᛬", "᠃", "᠉", "⁚", "∶", "ꓽ", "꞉", "︰", "：", ";", ";"])
// }, {
// 	character: "/",
// 	matcher: f(["᜵", "⁁", "⁄", "∕", "╱", "⟋", "⧸", "Ⳇ", "⼃", "〳", "ノ", "㇓", "丿", "\uD834\uDE3A"])
// }];
