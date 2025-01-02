# Fastly's Next-Gen WAF Terraform Configuration generation tool

This tool imports your NGWAF settings to the Terraform state and generates the Terraform configuration for your NGWAF config in HCL.

# TODO
Fix bug! Need to account for the case where the Terraform state is stored
remotely. Run a command like the following to generate the state file.
`terraform state pull > terraform.tfstate`

## Feature list and status
- [x] Corp Rules                
- [x] Corp Lists                
- [x] Corp Signals              
- [x] Site Request Rules        
- [x] Site Rate Limiting Rules  
- [x] Site Templated Rules      
- [x] Site Lists                
- [x] Site Signals              
- [x] Site Alerts               
- [x] Site Agent Alerts         
- [x] Site Integrations
- [ ] Header Link
- [ ] Edge integration
- [ ] Space lasers
- [ ] Coffee maker

# Set up
Environment variables must exist
```
TF_VAR_NGWAF_CORP
TF_VAR_NGWAF_EMAIL
TF_VAR_NGWAF_TOKEN
SIGSCI_CORP
SIGSCI_EMAIL
SIGSCI_TOKEN
```

Just run `make run`


# Need to start over?
`make rerun`

