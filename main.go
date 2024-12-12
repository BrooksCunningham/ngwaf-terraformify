package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
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

	existing_terraform_ids, err := ExtractTerraformStateIDs(
		filepath.Join(".", "terraform.tfstate"),
		"",
	)
	if err != nil {
		fmt.Println(err)
	}

	// Corp imports
	allCorpRules, _ := sc.GetAllCorpRules(corp)
	set_import_corp_rule_resources(allCorpRules, existing_terraform_ids)

	allCorpLists, _ := sc.GetAllCorpLists(corp)
	set_import_corp_list_resources(allCorpLists, existing_terraform_ids)

	allCorpSignals, _ := sc.GetAllCorpSignalTags(corp)
	set_import_corp_signals_resources(allCorpSignals, existing_terraform_ids)

	allSiteNames, _ := sc.ListSites(corp)
	set_import_sites_resources(allSiteNames, existing_terraform_ids)

	// Site imports
	for _, ngwafSite := range allSiteNames {
		// Site rules
		allSiteRules, _ := sc.GetAllSiteRules(corp, ngwafSite.Name)
		set_import_site_rule_resources(ngwafSite.Name, allSiteRules, existing_terraform_ids)

		// Site Legacy Templated Rules
		allLegacyTemplatedRules := get_active_legacy_templated_rules(sc, corp, ngwafSite.Name, email, token)
		set_import_site_legacy_templated_rule_resources(ngwafSite.Name, allLegacyTemplatedRules, existing_terraform_ids)

		// Site tags
		allSiteSignals, _ := sc.GetAllSiteSignalTags(corp, ngwafSite.Name)
		set_import_site_signals_resources(ngwafSite.Name, allSiteSignals, existing_terraform_ids)

		// Site lists
		allSiteLists, _ := sc.GetAllSiteLists(corp, ngwafSite.Name)
		set_import_site_list_resources(ngwafSite.Name, allSiteLists, existing_terraform_ids)

		allSiteIntegrations, _ := sc.ListIntegrations(corp, ngwafSite.Name)
		set_import_site_integration_resources(ngwafSite.Name, allSiteIntegrations, existing_terraform_ids)

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
		set_import_site_agent_alerts_resources(ngwafSite.Name, agentAlerts, existing_terraform_ids)
		set_import_site_alerts_resources(ngwafSite.Name, infoAlerts, existing_terraform_ids)
	}

	fmt.Println("done")

}

func set_import_corp_rule_resources(allCorpRules sigsci.ResponseCorpRuleBodyList, existing_terraform_ids []string) []string {
	var sigsciCorpIdNoNnumbersArray []string

	// Create a new empty HCL file
	file := hclwrite.NewEmptyFile()

	for _, corp_rule := range allCorpRules.Data {
		if slices.Contains(existing_terraform_ids, corp_rule.ID) {
			continue
		}
		sigsciCorpIdNoNnumbers := sanitizeTfId(corp_rule.ID)
		if corp_rule.Type == "request" {
			// Create a new block (e.g., a resource block)
			block := file.Body().AppendNewBlock("import", nil)
			// Set attributes for the block
			block.Body().SetAttributeValue("id", cty.StringVal(corp_rule.ID))
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
func set_import_corp_list_resources(list sigsci.ResponseListBodyList, existing_terraform_ids []string) []string {
	var sigsciIdNoNnumbersArray []string

	// Create a new empty HCL file
	file := hclwrite.NewEmptyFile()

	for _, item := range list.Data {
		if slices.Contains(existing_terraform_ids, item.ID) {
			continue
		}
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
func set_import_corp_signals_resources(allCorpList sigsci.ResponseSignalTagBodyList, existing_terraform_ids []string) []string {
	var sigsciIdNoNnumbersArray []string

	// Create a new empty HCL file
	file := hclwrite.NewEmptyFile()

	for _, item := range allCorpList.Data {
		if slices.Contains(existing_terraform_ids, item.TagName) {
			continue
		}
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
func set_import_sites_resources(allCorpList []sigsci.Site, existing_terraform_ids []string) []string {
	var sigsciIdNoNnumbersArray []string

	// Create a new empty HCL file
	file := hclwrite.NewEmptyFile()

	for _, item := range allCorpList {
		if slices.Contains(existing_terraform_ids, item.Name) {
			continue
		}
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
func set_import_site_list_resources(ngwafSiteShortName string, list sigsci.ResponseListBodyList, existing_terraform_ids []string) []string {
	var sigsciIdNoNnumbersArray []string

	// Create a new empty HCL file
	file := hclwrite.NewEmptyFile()

	for _, item := range list.Data {
		if slices.Contains(existing_terraform_ids, item.ID) {
			continue
		}
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
func set_import_site_integration_resources(ngwafSiteShortName string, list []sigsci.Integration, existing_terraform_ids []string) []string {
	var sigsciIdNoNnumbersArray []string

	// Create a new empty HCL file
	file := hclwrite.NewEmptyFile()

	for _, item := range list {
		if slices.Contains(existing_terraform_ids, item.ID) {
			continue
		}
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
func set_import_site_alerts_resources(ngwafSiteShortName string, list []sigsci.CustomAlert, existing_terraform_ids []string) []string {
	var sigsciIdNoNnumbersArray []string

	// Create a new empty HCL file
	file := hclwrite.NewEmptyFile()

	for _, item := range list {
		if slices.Contains(existing_terraform_ids, item.ID) {
			continue
		}
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
func set_import_site_agent_alerts_resources(ngwafSiteShortName string, list []sigsci.CustomAlert, existing_terraform_ids []string) []string {
	var sigsciIdNoNnumbersArray []string

	// Create a new empty HCL file
	file := hclwrite.NewEmptyFile()

	for _, item := range list {
		if slices.Contains(existing_terraform_ids, item.ID) {
			continue
		}
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

func set_import_site_signals_resources(ngwafSiteShortName string, list sigsci.ResponseSignalTagBodyList, existing_terraform_ids []string) []string {
	var sigsciIdNoNnumbersArray []string

	// Create a new empty HCL file
	file := hclwrite.NewEmptyFile()

	for _, item := range list.Data {
		if slices.Contains(existing_terraform_ids, item.TagName) {
			continue
		}
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

func set_import_site_rule_resources(ngwafSiteShortName string, allSiteRules sigsci.ResponseSiteRuleBodyList, existing_terraform_ids []string) []string {
	var sigsciSiteIdNoNnumbersArray []string

	// Create a new empty HCL file
	file := hclwrite.NewEmptyFile()

	for _, item := range allSiteRules.Data {
		if slices.Contains(existing_terraform_ids, item.ID) {
			continue
		}
		sigsciSiteIdNoNnumbers := sanitizeTfId(item.ID)

		switch item.Type {
		case "request":
			// Create a new block (e.g., a resource block)
			block := file.Body().AppendNewBlock("import", nil)
			// Set attributes for the block
			block.Body().SetAttributeValue("id", cty.StringVal(fmt.Sprintf(`%s:%s`, ngwafSiteShortName, item.ID)))
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
			block.Body().SetAttributeValue("id", cty.StringVal(fmt.Sprintf(`%s:%s`, ngwafSiteShortName, item.ID)))
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
			block.Body().SetAttributeValue("id", cty.StringVal(fmt.Sprintf(`%s:%s`, ngwafSiteShortName, item.ID)))
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

func set_import_site_legacy_templated_rule_resources(ngwafSiteShortName string, list ResponseSiteLegacyTemplatedRuleBodyList, existing_terraform_ids []string) []string {
	var sigsciIdNoNnumbersArray []string

	// Create a new empty HCL file
	file := hclwrite.NewEmptyFile()

	for _, item := range list.Data {
		if slices.Contains(existing_terraform_ids, item.Name) {
			continue
		}
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

// TerraformState represents the complete structure of a Terraform state file
type TerraformState struct {
	Version          int             `json:"version"`
	TerraformVersion string          `json:"terraform_version"`
	Serial           int             `json:"serial"`
	Lineage          string          `json:"lineage"`
	Resources        []ResourceState `json:"resources"`
}

// ResourceState represents a single resource in the Terraform state
type ResourceState struct {
	Type      string          `json:"type"`
	Name      string          `json:"name"`
	Provider  string          `json:"provider"`
	Instances []InstanceState `json:"instances"`
}

// InstanceState represents an instance of a resource
type InstanceState struct {
	SchemaVersion  int                    `json:"schema_version"`
	Attributes     map[string]interface{} `json:"attributes"`
	SensitiveAttrs []interface{}          `json:"sensitive_attributes,omitempty"`
	Private        string                 `json:"private,omitempty"`
}

// StateIDExtractor handles extraction of resource IDs from Terraform state
// type StateIDExtractor struct {
// 	statePath string
// 	state     *TerraformState
// }

// ExtractTerraformStateIDs consolidates file reading and ID extraction into a single function
func ExtractTerraformStateIDs(statePath string, resourceType string) ([]string, error) {
	if _, err := os.Stat(statePath); err == nil {
		// Read the file contents
		content, err := os.ReadFile(statePath)
		if err != nil {
			return nil, fmt.Errorf("error reading state file: %v", err)
		}

		// Parse the state file
		var state TerraformState
		if err := json.Unmarshal(content, &state); err != nil {
			return nil, fmt.Errorf("error parsing state file: %v", err)
		}

		// Extract IDs
		var ids []string
		for _, resource := range state.Resources {
			if resourceType == "" || resource.Type == resourceType {
				for _, instance := range resource.Instances {
					if id, ok := instance.Attributes["id"].(string); ok && id != "" {
						ids = append(ids, id)
					}
				}
			}
		}

		if len(ids) == 0 {
			return nil, fmt.Errorf("no IDs found")
		}

		return ids, nil
	}
	return nil, fmt.Errorf("no terraform.tfstate file found")
}
