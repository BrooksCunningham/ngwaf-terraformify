package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-exec/tfexec"
	sigsci "github.com/signalsciences/go-sigsci"
)

func main() {
	email := os.Getenv("TF_VAR_NGWAF_EMAIL")
	token := os.Getenv("TF_VAR_NGWAF_TOKEN")
	sc := sigsci.NewTokenClient(email, token)

	corp := os.Getenv("TF_VAR_NGWAF_CORP")
	site := os.Getenv("TF_VAR_NGWAF_SITE")

	// Step 1
	execPath := "/usr/local/bin/terraform"
	workingDir, _ := os.Getwd()

	// perform basic terraform file setup
	set_up_providers(workingDir)
	set_up_versions(workingDir)
	set_up_tf_variables(workingDir)
	set_up_tf_import(workingDir)

	// perform basic terraform init
	tf, err := tfexec.NewTerraform(workingDir, execPath)
	if err != nil {
		log.Fatalf("error running NewTerraform: %s", err)
	}

	err = tf.Init(context.Background(), tfexec.Upgrade(true))
	if err != nil {
		log.Fatalf("error running Init: %s", err)
	}

	// Get all site rules for a site.
	allSiteRules, err := sc.GetAllSiteRules(corp, site)
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.OpenFile(workingDir+`/import.tf`, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		log.Fatalf("Error writing to import.tf: %s", err)
	}
	defer f.Close()

	for _, siteRule := range allSiteRules.Data {
		// fmt.Println("At index", index, "value is", siteRule)
		sigsciSiteIdNoNnumbers := removeDigits(siteRule.ID)
		siteImportData := []byte(fmt.Sprintf(`
import {
  to = sigsci_site_rule.%v
  id = "terraform_ngwaf_site:%v"
}
		`, sigsciSiteIdNoNnumbers, siteRule.ID))
		_, err := f.Write(siteImportData)
		if err != nil {
			log.Fatal(err)
			return
		}
	}
}

func import_site_resources(tf *tfexec.Terraform, context context.Context, site string, sigsciSiteRules []sigsci.ResponseSiteRuleBody) []string {
	var sigsciSiteIdNoNnumbersArray []string
	for _, siteRule := range sigsciSiteRules {
		sigsciSiteIdNoNnumbers := removeDigits(siteRule.ID)
		if siteRule.Type == "request" {
			// siteRuleJson, _ := json.MarshalIndent(siteRule, "", "    ")
			// log.Printf(`Importing: %s`, siteRule.ID)
			// terraform import sigsci_site_rule.ebcdceed terraform_ngwaf_site:64de89736993ba01d4fc06ba
			err := tf.Import(context, fmt.Sprintf(`%s.%s`, `sigsci_site_rule`, sigsciSiteIdNoNnumbers), fmt.Sprintf("%s:%s", site, siteRule.ID))
			if err != nil {
				log.Fatalf("error running Import: %s", err)
			}
			sigsciSiteIdNoNnumbersArray = append(sigsciSiteIdNoNnumbersArray, sigsciSiteIdNoNnumbers)
		}
	}
	return sigsciSiteIdNoNnumbersArray
}

func add_site_resources(f *os.File, sigsciSiteRules []sigsci.ResponseSiteRuleBody) {

	for _, siteRule := range sigsciSiteRules {
		if siteRule.Type == "request" {
			sigsciSiteIdNoNnumbers := removeDigits(siteRule.ID)

			data := fmt.Sprintf(`resource "sigsci_site_rule" "%s" {}
		`, sigsciSiteIdNoNnumbers)
			f.WriteString(data)
		}
	}
	f.Sync()
}

func set_up_versions(workingDir string) {
	d1 := []byte(`
# Terraform 0.13+ requires providers to be declared in a "required_providers" block
# https://registry.terraform.io/providers/fastly/fastly/latest/docs
terraform {
  required_providers {
    sigsci = {
      source = "signalsciences/sigsci"
      version = ">= 2.1.0"
    }
  }
}
	`)
	err := os.WriteFile(workingDir+`/versions.tf`, d1, 0644)
	if err != nil {
		log.Fatalf("Error writing: %s", err)
	}
}

func set_up_providers(workingDir string) {
	d1 := []byte(`
provider "sigsci" {
	corp = var.NGWAF_CORP
	email = var.NGWAF_EMAIL
	auth_token = var.NGWAF_TOKEN
}
	`)
	err := os.WriteFile(workingDir+`/providers.tf`, d1, 0644)
	if err != nil {
		log.Fatalf("Error writing: %s", err)
	}
}

func set_up_tf_variables(workingDir string) {
	d1 := []byte(`
#### NGWAF variables - Start

variable "NGWAF_CORP" {
	type          = string
	description   = "Corp name for NGWAF"
}

variable "NGWAF_SITE" {
	type          = string
	description   = "Site name for NGWAF"
}

variable "NGWAF_EMAIL" {
	type        = string
	description = "Email address associated with the token for the NGWAF API."
}
variable "NGWAF_TOKEN" {
	type        = string
	description = "Secret token for the NGWAF API."
	sensitive   = true
}
#### NGWAF variables - End
	`)
	err := os.WriteFile(workingDir+`/variables.tf`, d1, 0644)
	if err != nil {
		log.Fatalf("Error writing: %s", err)
	}
}

func set_up_tf_import(workingDir string) {
	d1 := []byte(``)
	err := os.WriteFile(workingDir+`/import.tf`, d1, 0644)
	if err != nil {
		log.Fatalf("Error writing: %s", err)
	}
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
