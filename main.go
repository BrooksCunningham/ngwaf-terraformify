package main

import (
	"fmt"
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
	// site := os.Getenv("TF_VAR_NGWAF_SITE")

	// Corp imports
	allCorpRules, _ := sc.GetAllCorpRules(corp)
	set_import_corp_rule_resources(allCorpRules)

	allCorpLists, _ := sc.GetAllCorpLists(corp)
	set_import_corp_list_resources(allCorpLists)

	allCorpSignals, _ := sc.GetAllCorpSignalTags(corp)
	set_import_corp_signals_resources(allCorpSignals)

	allSiteNames, _ := sc.ListSites(corp)
	set_import_sites_resources(allSiteNames)
	// fmt.Println(allSiteNames)

	for _, ngwafSite := range allSiteNames {
		// Site imports
		allSiteRules, _ := sc.GetAllSiteRules(corp, ngwafSite.Name)
		set_import_site_rule_resources(ngwafSite.Name, allSiteRules)

		allSiteSignals, _ := sc.GetAllSiteSignalTags(corp, ngwafSite.Name)
		set_import_site_signals_resources(ngwafSite.Name, allSiteSignals)

		allSiteLists, _ := sc.GetAllSiteLists(corp, ngwafSite.Name)
		set_import_site_list_resources(ngwafSite.Name, allSiteLists)
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
	fileImportTf, _ := os.OpenFile("import.tf", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if _, err := file.WriteTo(fileImportTf); err != nil {
		fmt.Println(`Error writing HCL:`, err)
		os.Exit(1)
	}
	defer fileImportTf.Close()
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
		// }
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
		// }
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
	fileImportTf, _ := os.OpenFile("import.tf", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if _, err := file.WriteTo(fileImportTf); err != nil {
		fmt.Println(`Error writing HCL:`, err)
		os.Exit(1)
	}
	defer fileImportTf.Close()
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
	fileImportTf, _ := os.OpenFile("import.tf", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if _, err := file.WriteTo(fileImportTf); err != nil {
		fmt.Println(`Error writing HCL:`, err)
		os.Exit(1)
	}
	defer fileImportTf.Close()
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
	fileImportTf, _ := os.OpenFile("import.tf", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if _, err := file.WriteTo(fileImportTf); err != nil {
		fmt.Println(`Error writing HCL:`, err)
		os.Exit(1)
	}
	defer fileImportTf.Close()
	return sigsciIdNoNnumbersArray
}

func set_import_site_rule_resources(ngwafSiteShortName string, allSiteRules sigsci.ResponseSiteRuleBodyList) []string {
	var sigsciSiteIdNoNnumbersArray []string

	// Create a new empty HCL file
	file := hclwrite.NewEmptyFile()

	for _, siteRule := range allSiteRules.Data {
		sigsciSiteIdNoNnumbers := sanitizeTfId(siteRule.ID)
		if siteRule.Type == "request" {
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
	fileImportTf, _ := os.OpenFile("import.tf", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if _, err := file.WriteTo(fileImportTf); err != nil {
		fmt.Println(`Error writing HCL:`, err)
		os.Exit(1)
	}
	defer fileImportTf.Close()
	return sigsciSiteIdNoNnumbersArray
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
