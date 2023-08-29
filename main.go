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
	email := os.Getenv("NGWAF_EMAIL")
	token := os.Getenv("NGWAF_TOKEN")
	sc := sigsci.NewTokenClient(email, token)

	corp := os.Getenv("NGWAF_CORP")
	site := os.Getenv("TF_VAR_NGWAF_SITE")
	allSiteRules, err := sc.GetAllSiteRules(corp, site)
	if err != nil {
		log.Fatal(err)
	}

	// TODOs
	// ##### 1
	// Perform tf import
	// https://registry.terraform.io/providers/signalsciences/sigsci/latest/docs/resources/site
	// terraform import sigsci_site_rule.test id
	// terraform import sigsci_site_rule.64de89736993ba01d4fc06ba 64de89736993ba01d4fc06ba
	// terraform import sigsci_site_rule.ebcdceed terraform_ngwaf_site:64de89736993ba01d4fc06ba
	// ##### 2
	// parse the imported tf state to build the configuration
	// terraform state show sigsci_site.test

	// Step 1
	execPath := "/usr/local/bin/terraform"
	workingDir, _ := os.Getwd()

	log.Println(execPath)
	log.Println(workingDir)

	// perform basic terraform file setup
	set_up_providers(workingDir)
	set_up_versions(workingDir)
	set_up_tf_variables(workingDir)
	set_up_tf_imports(workingDir)

	// https://gobyexample.com/writing-files
	// f, err := os.Create(workingDir + `/ngwaf.tf`)
	// check(err)
	// defer f.Close()

	// iterate over site rules and write a placeholder for a future terraform import
	// add_site_resources(f, allSiteRules.Data)

	// perform basic terraform init
	tf, err := tfexec.NewTerraform(workingDir, execPath)
	if err != nil {
		log.Fatalf("error running NewTerraform: %s", err)
	}

	err = tf.Init(context.Background(), tfexec.Upgrade(true))
	if err != nil {
		log.Fatalf("error running Init: %s", err)
	}

	// Import the site rules
	import_site_resources(tf, context.Background(), site, allSiteRules.Data)

	// https://github.com/hashicorp/terraform-exec/issues/336 :-(
	// state, err := tf.Show(context.Background())
	// if err != nil {
	// 	log.Fatalf("error running Show: %s", err)
	// }

	// log.Println(state.Values)
	// log.Println(state.Values.Outputs)
	// log.Println(state.Values.RootModule.Resources)
	// log.Println()
	// siteRuleJson, _ := json.MarshalIndent(state.Values.RootModule.Resources[0], "", "    ")
	// log.Println(string(siteRuleJson))

	// log.Println(state.Values.RootModule.Address)
}

func import_site_resources(tf *tfexec.Terraform, context context.Context, site string, sigsciSiteRules []sigsci.ResponseSiteRuleBody) []string {
	var sigsciSiteIdNoNnumbersArray []string
	for _, siteRule := range sigsciSiteRules {
		sigsciSiteIdNoNnumbers := removeDigits(siteRule.ID)
		if siteRule.Type == "request" {
			// siteRuleJson, _ := json.MarshalIndent(siteRule, "", "    ")
			log.Printf(`Importing: %s`, siteRule.ID)
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

// err = tf.Import(context.Background(), "sigsci_site.64de89736993ba01d4fc06ba", "64de89736993ba01d4fc06ba")
func add_site_resources(f *os.File, sigsciSiteRules []sigsci.ResponseSiteRuleBody) {

	for _, siteRule := range sigsciSiteRules {
		// log.Println(siteRule)
		if siteRule.Type == "request" {
			// siteRuleJson, _ := json.MarshalIndent(siteRule, "", "    ")
			// log.Println(string(siteRuleJson))
			log.Println(siteRule.ID)
			log.Println("write resource to file ", siteRule.ID)

			sigsciSiteIdNoNnumbers := removeDigits(siteRule.ID)

			data := fmt.Sprintf(`resource "sigsci_site_rule" "%s" {}
		`, sigsciSiteIdNoNnumbers)
			f.WriteString(data)
		}
	}
	// fmt.Println(err)
	// fmt.Printf("wrote %d bytes\n", n3)
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
	log.Println(err)
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
	log.Println(err)
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
	log.Println(err)
}

func set_up_tf_imports(workingDir string) {
	d1 := []byte(`
	import {
		to = sigsci_site_rule.ebcdceed
		id = "terraform_ngwaf_site:64de89736993ba01d4fc06ba"
	  }
	  
	`)
	err := os.WriteFile(workingDir+`/import.tf`, d1, 0644)
	log.Println(err)
}

func removeDigits(str string) string {
	var sb strings.Builder
	chars := []rune(str)
	for i := 0; i < len(chars); i++ {
		char := string(chars[i])
		sb.WriteString(toCharStrConst(char))

	}
	log.Println(sb.String())

	return sb.String()
}

func toCharStrConst(s string) string {

	if sNum, err := strconv.ParseInt(s, 10, 32); err == nil {
		//		fmt.Printf("%T, %v\n", s, sNum)
		const abc = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
		sNum += 1 // add plus one to handle int 0 (zero)
		log.Println(sNum)
		return abc[sNum-1 : sNum]
	}
	return s
}
