package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	sigsci "github.com/signalsciences/go-sigsci"
	"github.com/zclconf/go-cty/cty"
)

func main() {
	email := os.Getenv("TF_VAR_NGWAF_EMAIL")
	token := os.Getenv("TF_VAR_NGWAF_TOKEN")
	sc := sigsci.NewTokenClient(email, token)

	corp := os.Getenv("TF_VAR_NGWAF_CORP")

	// Corp imports
	allCorpRules, _ := sc.GetAllCorpRules(corp)
	set_import_corp_rule_resources(allCorpRules)

	allCorpLists, _ := sc.GetAllCorpLists(corp)
	set_import_corp_list_resources(allCorpLists)

	allCorpSignals, _ := sc.GetAllCorpSignalTags(corp)
	set_import_corp_signals_resources(allCorpSignals)

	allSiteNames, _ := sc.ListSites(corp)
	set_import_sites_resources(allSiteNames)

	// Site imports
	for _, ngwafSite := range allSiteNames {
		// Site rules
		allSiteRules, _ := sc.GetAllSiteRules(corp, ngwafSite.Name)
		set_import_site_rule_resources(ngwafSite.Name, allSiteRules)

		// Site Legacy Templated Rules
		allLegacyTemplatedRules := get_active_legacy_templated_rules(sc, corp, ngwafSite.Name, email, token)
		set_import_site_legacy_templated_rule_resources(ngwafSite.Name, allLegacyTemplatedRules)

		// Site tags
		allSiteSignals, _ := sc.GetAllSiteSignalTags(corp, ngwafSite.Name)
		set_import_site_signals_resources(ngwafSite.Name, allSiteSignals)

		// Site lists
		allSiteLists, _ := sc.GetAllSiteLists(corp, ngwafSite.Name)
		set_import_site_list_resources(ngwafSite.Name, allSiteLists)

		allSiteIntegrations, _ := sc.ListIntegrations(corp, ngwafSite.Name)
		set_import_site_integration_resources(ngwafSite.Name, allSiteIntegrations)

		// Site alerts and Agent alerts
		allSiteAlerts, _ := sc.ListCustomAlerts(corp, ngwafSite.Name)

		var infoAlerts []sigsci.CustomAlert
		var agentAlerts []sigsci.CustomAlert
		for _, siteAlert := range allSiteAlerts {
			if siteAlert.Action == "info" {
				infoAlerts = append(infoAlerts, siteAlert)
			}
			if siteAlert.Action == "siteMetricInfo" {
				agentAlerts = append(agentAlerts, siteAlert)
			}
		}

		// Site agent alerts and Site alerts
		set_import_site_agent_alerts_resources(ngwafSite.Name, agentAlerts)
		set_import_site_alerts_resources(ngwafSite.Name, infoAlerts)
	}

	fmt.Println("done")

}

func set_import_corp_rule_resources(allCorpRules sigsci.ResponseCorpRuleBodyList) []string {
	var sigsciCorpIdNoNnumbersArray []string

	// Create a new empty HCL file
	file := hclwrite.NewEmptyFile()

	for _, corpRule := range allCorpRules.Data {
		sigsciCorpIdNoNnumbers := sanitizeTfId(corpRule.ID)
		if corpRule.Type == "request" {
			// Create a new block (e.g., a resource block)
			block := file.Body().AppendNewBlock("import", nil)
			// Set attributes for the block
			block.Body().SetAttributeValue("id", cty.StringVal(corpRule.ID))
			tokens := hclwrite.Tokens{
				{
					Type:  hclsyntax.TokenIdent,
					Bytes: []byte(fmt.Sprintf(`sigsci_corp_rule.%s`, sigsciCorpIdNoNnumbers)),
				},
			}
			block.Body().SetAttributeRaw("to", tokens)
			sigsciCorpIdNoNnumbersArray = append(sigsciCorpIdNoNnumbersArray, sigsciCorpIdNoNnumbers)
		}
	}

	// Open the file and write
	write_terraform_config_to_file(file, "import.tf")
	return sigsciCorpIdNoNnumbersArray
}

// Corp lists
func set_import_corp_list_resources(list sigsci.ResponseListBodyList) []string {
	var sigsciIdNoNnumbersArray []string

	// Create a new empty HCL file
	file := hclwrite.NewEmptyFile()

	for _, item := range list.Data {
		sigsciIdNoNnumbers := sanitizeTfId(item.ID)
		// if item.Type == "request" {
		// Create a new block (e.g., a resource block)
		block := file.Body().AppendNewBlock("import", nil)
		// Set attributes for the block
		block.Body().SetAttributeValue("id", cty.StringVal(item.ID))
		tokens := hclwrite.Tokens{
			{
				Type:  hclsyntax.TokenIdent,
				Bytes: []byte(fmt.Sprintf(`sigsci_corp_list.%s`, sigsciIdNoNnumbers)),
			},
		}
		block.Body().SetAttributeRaw("to", tokens)
		sigsciIdNoNnumbersArray = append(sigsciIdNoNnumbersArray, sigsciIdNoNnumbers)
	}

	// Open the file and write
	fileImportTf, _ := os.OpenFile("import.tf", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if _, err := file.WriteTo(fileImportTf); err != nil {
		fmt.Println(`Error writing HCL:`, err)
		os.Exit(1)
	}
	defer fileImportTf.Close()
	return sigsciIdNoNnumbersArray
}

// Corp Signals
func set_import_corp_signals_resources(allCorpList sigsci.ResponseSignalTagBodyList) []string {
	var sigsciIdNoNnumbersArray []string

	// Create a new empty HCL file
	file := hclwrite.NewEmptyFile()

	for _, item := range allCorpList.Data {
		sigsciIdNoNnumbers := sanitizeTfId(item.TagName)
		block := file.Body().AppendNewBlock("import", nil)
		// Set attributes for the block
		block.Body().SetAttributeValue("id", cty.StringVal(item.TagName))
		tokens := hclwrite.Tokens{
			{
				Type:  hclsyntax.TokenIdent,
				Bytes: []byte(fmt.Sprintf(`sigsci_corp_signal_tag.%s`, sigsciIdNoNnumbers)),
			},
		}
		block.Body().SetAttributeRaw("to", tokens)
		sigsciIdNoNnumbersArray = append(sigsciIdNoNnumbersArray, sigsciIdNoNnumbers)
	}

	// Open the file and write
	write_terraform_config_to_file(file, "import.tf")
	return sigsciIdNoNnumbersArray
}

// Sites
func set_import_sites_resources(allCorpList []sigsci.Site) []string {
	var sigsciIdNoNnumbersArray []string

	// Create a new empty HCL file
	file := hclwrite.NewEmptyFile()

	for _, item := range allCorpList {
		sigsciIdNoNnumbers := sanitizeTfId(item.Name)
		block := file.Body().AppendNewBlock("import", nil)
		// Set attributes for the block
		block.Body().SetAttributeValue("id", cty.StringVal(item.Name))
		tokens := hclwrite.Tokens{
			{
				Type:  hclsyntax.TokenIdent,
				Bytes: []byte(fmt.Sprintf(`sigsci_site.%s`, sigsciIdNoNnumbers)),
			},
		}
		block.Body().SetAttributeRaw("to", tokens)
		sigsciIdNoNnumbersArray = append(sigsciIdNoNnumbersArray, sigsciIdNoNnumbers)
	}

	// Open the file and write
	write_terraform_config_to_file(file, "import.tf")

	return sigsciIdNoNnumbersArray
}

// Site lists
func set_import_site_list_resources(ngwafSiteShortName string, list sigsci.ResponseListBodyList) []string {
	var sigsciIdNoNnumbersArray []string

	// Create a new empty HCL file
	file := hclwrite.NewEmptyFile()

	for _, item := range list.Data {
		sigsciIdNoNnumbers := sanitizeTfId(item.ID)
		// Create a new block (e.g., a resource block)
		block := file.Body().AppendNewBlock("import", nil)
		// Set attributes for the block
		block.Body().SetAttributeValue("id", cty.StringVal(fmt.Sprintf(`%s:%s`, ngwafSiteShortName, item.ID)))
		tokens := hclwrite.Tokens{
			{
				Type:  hclsyntax.TokenIdent,
				Bytes: []byte(fmt.Sprintf(`sigsci_site_list.%s%s`, ngwafSiteShortName, sigsciIdNoNnumbers)),
			},
		}
		block.Body().SetAttributeRaw("to", tokens)
		sigsciIdNoNnumbersArray = append(sigsciIdNoNnumbersArray, sigsciIdNoNnumbers)
	}

	// Open the file and write
	write_terraform_config_to_file(file, "import.tf")
	return sigsciIdNoNnumbersArray
}

// Site alerts
func set_import_site_integration_resources(ngwafSiteShortName string, list []sigsci.Integration) []string {
	var sigsciIdNoNnumbersArray []string

	// Create a new empty HCL file
	file := hclwrite.NewEmptyFile()

	for _, item := range list {
		sigsciIdNoNnumbers := sanitizeTfId(item.ID)
		// Create a new block (e.g., a resource block)
		block := file.Body().AppendNewBlock("import", nil)
		// Set attributes for the block
		block.Body().SetAttributeValue("id", cty.StringVal(fmt.Sprintf(`%s:%s`, ngwafSiteShortName, item.ID)))
		tokens := hclwrite.Tokens{
			{
				Type:  hclsyntax.TokenIdent,
				Bytes: []byte(fmt.Sprintf(`sigsci_site_integration.%s`, sigsciIdNoNnumbers)),
			},
		}
		block.Body().SetAttributeRaw("to", tokens)
		sigsciIdNoNnumbersArray = append(sigsciIdNoNnumbersArray, sigsciIdNoNnumbers)
	}

	// Open the file and write
	write_terraform_config_to_file(file, "import.tf")
	return sigsciIdNoNnumbersArray
}

// Site alerts
func set_import_site_alerts_resources(ngwafSiteShortName string, list []sigsci.CustomAlert) []string {
	var sigsciIdNoNnumbersArray []string

	// Create a new empty HCL file
	file := hclwrite.NewEmptyFile()

	for _, item := range list {
		sigsciIdNoNnumbers := sanitizeTfId(item.ID)
		// Create a new block (e.g., a resource block)
		block := file.Body().AppendNewBlock("import", nil)
		// Set attributes for the block
		block.Body().SetAttributeValue("id", cty.StringVal(fmt.Sprintf(`%s:%s`, ngwafSiteShortName, item.ID)))
		tokens := hclwrite.Tokens{
			{
				Type:  hclsyntax.TokenIdent,
				Bytes: []byte(fmt.Sprintf(`sigsci_site_alert.%s`, sigsciIdNoNnumbers)),
			},
		}
		block.Body().SetAttributeRaw("to", tokens)
		sigsciIdNoNnumbersArray = append(sigsciIdNoNnumbersArray, sigsciIdNoNnumbers)
	}

	// Open the file and write
	write_terraform_config_to_file(file, "import.tf")
	return sigsciIdNoNnumbersArray
}

// Site agent alerts
func set_import_site_agent_alerts_resources(ngwafSiteShortName string, list []sigsci.CustomAlert) []string {
	var sigsciIdNoNnumbersArray []string

	// Create a new empty HCL file
	file := hclwrite.NewEmptyFile()

	for _, item := range list {
		sigsciIdNoNnumbers := sanitizeTfId(item.ID)
		// Create a new block (e.g., a resource block)
		block := file.Body().AppendNewBlock("import", nil)
		// Set attributes for the block
		block.Body().SetAttributeValue("id", cty.StringVal(fmt.Sprintf(`%s:%s`, ngwafSiteShortName, item.ID)))
		tokens := hclwrite.Tokens{
			{
				Type:  hclsyntax.TokenIdent,
				Bytes: []byte(fmt.Sprintf(`sigsci_site_agent_alert.%s`, sigsciIdNoNnumbers)),
			},
		}
		block.Body().SetAttributeRaw("to", tokens)
		sigsciIdNoNnumbersArray = append(sigsciIdNoNnumbersArray, sigsciIdNoNnumbers)
	}

	// Open the file and write
	write_terraform_config_to_file(file, "import.tf")
	return sigsciIdNoNnumbersArray
}

func set_import_site_signals_resources(ngwafSiteShortName string, list sigsci.ResponseSignalTagBodyList) []string {
	var sigsciIdNoNnumbersArray []string

	// Create a new empty HCL file
	file := hclwrite.NewEmptyFile()

	for _, item := range list.Data {
		sigsciIdNoNnumbers := sanitizeTfId(item.TagName)
		block := file.Body().AppendNewBlock("import", nil)
		// Set attributes for the block
		block.Body().SetAttributeValue("id", cty.StringVal(fmt.Sprintf(`%s:%s`, ngwafSiteShortName, item.TagName)))
		tokens := hclwrite.Tokens{
			{
				Type:  hclsyntax.TokenIdent,
				Bytes: []byte(fmt.Sprintf(`sigsci_site_signal_tag.%s%s`, ngwafSiteShortName, sigsciIdNoNnumbers)),
			},
		}
		block.Body().SetAttributeRaw("to", tokens)
		sigsciIdNoNnumbersArray = append(sigsciIdNoNnumbersArray, sigsciIdNoNnumbers)
		// }
	}

	// Open the file and write
	write_terraform_config_to_file(file, "import.tf")

	return sigsciIdNoNnumbersArray
}

func set_import_site_rule_resources(ngwafSiteShortName string, allSiteRules sigsci.ResponseSiteRuleBodyList) []string {
	var sigsciSiteIdNoNnumbersArray []string

	// Create a new empty HCL file
	file := hclwrite.NewEmptyFile()

	for _, siteRule := range allSiteRules.Data {
		sigsciSiteIdNoNnumbers := sanitizeTfId(siteRule.ID)

		switch siteRule.Type {
		case "request":
			// Create a new block (e.g., a resource block)
			block := file.Body().AppendNewBlock("import", nil)
			// Set attributes for the block
			block.Body().SetAttributeValue("id", cty.StringVal(fmt.Sprintf(`%s:%s`, ngwafSiteShortName, siteRule.ID)))
			tokens := hclwrite.Tokens{
				{
					Type:  hclsyntax.TokenIdent,
					Bytes: []byte(fmt.Sprintf(`sigsci_site_rule.%s`, sigsciSiteIdNoNnumbers)),
				},
			}
			block.Body().SetAttributeRaw("to", tokens)
			sigsciSiteIdNoNnumbersArray = append(sigsciSiteIdNoNnumbersArray, sigsciSiteIdNoNnumbers)
		case "rateLimit":
			// fmt.Printf("%+v\n", siteRule)
			// Create a new block (e.g., a resource block)
			block := file.Body().AppendNewBlock("import", nil)
			// Set attributes for the block
			block.Body().SetAttributeValue("id", cty.StringVal(fmt.Sprintf(`%s:%s`, ngwafSiteShortName, siteRule.ID)))
			tokens := hclwrite.Tokens{
				{
					Type:  hclsyntax.TokenIdent,
					Bytes: []byte(fmt.Sprintf(`sigsci_site_rule.%s`, sigsciSiteIdNoNnumbers)),
				},
			}
			block.Body().SetAttributeRaw("to", tokens)
			sigsciSiteIdNoNnumbersArray = append(sigsciSiteIdNoNnumbersArray, sigsciSiteIdNoNnumbers)

		case "templatedSignal":
			// Create a new block (e.g., a resource block)
			block := file.Body().AppendNewBlock("import", nil)
			// Set attributes for the block
			block.Body().SetAttributeValue("id", cty.StringVal(fmt.Sprintf(`%s:%s`, ngwafSiteShortName, siteRule.ID)))
			tokens := hclwrite.Tokens{
				{
					Type:  hclsyntax.TokenIdent,
					Bytes: []byte(fmt.Sprintf(`sigsci_site_rule.%s`, sigsciSiteIdNoNnumbers)),
				},
			}
			block.Body().SetAttributeRaw("to", tokens)
			sigsciSiteIdNoNnumbersArray = append(sigsciSiteIdNoNnumbersArray, sigsciSiteIdNoNnumbers)
		}
	}
	// Open the file and write
	write_terraform_config_to_file(file, "import.tf")

	return sigsciSiteIdNoNnumbersArray
}

func set_import_site_legacy_templated_rule_resources(ngwafSiteShortName string, list ResponseSiteLegacyTemplatedRuleBodyList) []string {
	var sigsciIdNoNnumbersArray []string

	// Create a new empty HCL file
	file := hclwrite.NewEmptyFile()

	for _, item := range list.Data {
		if len(item.Detections) > 0 {
			sigsciIdNoNnumbers := sanitizeTfId(item.Name)
			// Create a new block (e.g., a resource block)
			block := file.Body().AppendNewBlock("import", nil)
			// Set attributes for the block
			block.Body().SetAttributeValue("id", cty.StringVal(fmt.Sprintf(`%s:%s`, ngwafSiteShortName, item.Name)))
			tokens := hclwrite.Tokens{
				{
					Type:  hclsyntax.TokenIdent,
					Bytes: []byte(fmt.Sprintf(`sigsci_site_templated_rule.%s%s`, ngwafSiteShortName, sigsciIdNoNnumbers)),
				},
			}
			block.Body().SetAttributeRaw("to", tokens)
			sigsciIdNoNnumbersArray = append(sigsciIdNoNnumbersArray, sigsciIdNoNnumbers)
		}
	}

	// Open the file and write
	write_terraform_config_to_file(file, "import.tf")
	return sigsciIdNoNnumbersArray
}

func sanitizeTfId(str string) string {
	var sb strings.Builder
	chars := []rune(str)
	for i := 0; i < len(chars); i++ {
		if string(chars[i]) == "." {
			char := "dot"
			sb.WriteString(toCharStrConst(char))
		} else {
			char := string(chars[i])
			sb.WriteString(toCharStrConst(char))
		}
	}
	return sb.String()
}

func toCharStrConst(s string) string {

	if sNum, err := strconv.ParseInt(s, 10, 32); err == nil {
		const abc = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		sNum += 1 // add plus one to handle int 0 (zero)
		return abc[sNum-1 : sNum]
	}
	return s
}

func write_terraform_config_to_file(hclFile *hclwrite.File, fileName string) bool {

	// Open the file and write
	fileImportTf, _ := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if _, err := hclFile.WriteTo(fileImportTf); err != nil {
		fmt.Println(`Error writing HCL:`, err)
		return false
	}
	defer fileImportTf.Close()
	return true
}

func get_active_legacy_templated_rules(sc sigsci.Client, corpName string, siteName string, email string, token string) ResponseSiteLegacyTemplatedRuleBodyList {
	// Data structure for JSON input

	resp, _ := doRequestDetailed("GET", fmt.Sprintf("/v0/corps/%s/sites/%s/configuredtemplates", corpName, siteName), "", email, token)

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var legacyTemplatedRuledata ResponseSiteLegacyTemplatedRuleBodyList

	json.Unmarshal(body, &legacyTemplatedRuledata)

	return legacyTemplatedRuledata
}

func doRequestDetailed(method string, url string, reqBody string, email string, token string) (*http.Response, error) {
	apiURL := "https://dashboard.signalsciences.net/api"
	client := &http.Client{}

	var b io.Reader
	if reqBody != "" {
		b = strings.NewReader(reqBody)
	}

	req, _ := http.NewRequest(method, apiURL+url, b)

	if email != "" {
		// token auth
		req.Header.Set("X-API-User", email)
		req.Header.Set("X-API-Token", token)
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Set("User-Agent", "go-sigsci")

	resp, err := client.Do(req)

	return resp, err
}

type ResponseSiteLegacyTemplatedRuleBodyList struct {
	TotalCount int                                   `json:"totalCount"`
	Data       []ResponseSiteLegacyTemplatedRuleBody `json:"data"`
}

type ResponseSiteLegacyTemplatedRuleBody struct {
	CreateSiteLegacyTemplatedRuleBody
	Name       string      `json:"name"`
	Detections []Detection `json:"detections"`
}

type CreateSiteLegacyTemplatedRuleBody struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Enabled   bool    `json:"enabled"`
	Fields    []Field `json:"fields"`
	Created   string  `json:"created"`
	CreatedBy string  `json:"createdBy"`
}

type Field struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Detection struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Enabled   bool    `json:"enabled"`
	Fields    []Field `json:"fields"`
	Created   string  `json:"created"`
	CreatedBy string  `json:"createdBy"`
}
