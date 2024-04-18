package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	// "github.com/hashicorp/terraform-exec/tfexec"
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
	site := os.Getenv("TF_VAR_NGWAF_SITE")

	// Step 1
	// execPath := "/usr/local/bin/terraform"
	// workingDir, _ := os.Getwd()
	// fmt.Println("workingDir", workingDir)

	allSiteRules, _ := sc.GetAllSiteRules(corp, site)
	set_import_site_rule_resources(allSiteRules)

	allCorpRules, _ := sc.GetAllCorpRules(corp)
	set_import_corp_rule_resources(allCorpRules)

	// fmt.Println(allCorpRules)

	fmt.Println("done")

}

func set_import_corp_rule_resources(allCorpRules sigsci.ResponseCorpRuleBodyList) []string {
	var sigsciCorpIdNoNnumbersArray []string

	// Create a new empty HCL file
	file := hclwrite.NewEmptyFile()

	for _, corpRule := range allCorpRules.Data {
		// fmt.Println(`Importing:`, siteRule)
		sigsciCorpIdNoNnumbers := removeDigits(corpRule.ID)
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

func set_import_site_rule_resources(allSiteRules sigsci.ResponseSiteRuleBodyList) []string {
	var sigsciSiteIdNoNnumbersArray []string

	// Create a new empty HCL file
	file := hclwrite.NewEmptyFile()

	for _, siteRule := range allSiteRules.Data {
		// fmt.Println(`Importing:`, siteRule)
		sigsciSiteIdNoNnumbers := removeDigits(siteRule.ID)
		if siteRule.Type == "request" {
			// Create a new block (e.g., a resource block)
			block := file.Body().AppendNewBlock("import", nil)
			// Set attributes for the block
			block.Body().SetAttributeValue("id", cty.StringVal(fmt.Sprintf(`terraform_ngwaf_site:%s`, siteRule.ID)))
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

func removeDigits(str string) string {
	var sb strings.Builder
	chars := []rune(str)
	for i := 0; i < len(chars); i++ {
		char := string(chars[i])
		sb.WriteString(toCharStrConst(char))

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
